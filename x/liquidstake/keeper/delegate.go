package keeper

import (
	"time"

	"github.com/gogo/protobuf/proto"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

// Delegate performs a liquid stake delegation. delegator transfer the ibcToken to module account then
// get derivative token by the rate.
func (k *Keeper) Delegate(ctx sdk.Context, chainID string, amount math.Int, caller sdk.AccAddress) error {
	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", chainID)
	}

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, types.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", types.DelegationEpochIdentifier)
	}

	currentEpoch := uint64(epochInfo.CurrentEpoch)
	recordID, found := k.GetChianDelegationRecordID(ctx, chainID, currentEpoch)
	if !found {
		return sdkerrors.Wrapf(types.ErrNoExistDelegationRecord, "chainID %s, epoch %d", chainID, currentEpoch)
	}

	record, found := k.GetDelegationRecord(ctx, recordID)
	if !found {
		return sdkerrors.Wrapf(types.ErrNoExistDelegationRecord, "chainID %s, epoch %d, recorID %d", chainID, currentEpoch, recordID)
	}

	ecsrowAccAddress := sdk.MustAccAddressFromBech32(sourceChain.EcsrowAddress)
	// transfer ibc token to sourcechain's ecsrow account
	if err := k.sendCoinsFromAccountToAccount(ctx, caller,
		ecsrowAccAddress, sdk.Coins{sdk.NewCoin(sourceChain.IbcDenom, amount)}); err != nil {
		return err
	}

	// TODO replace TruncateInt with Ceil ?
	derivativeAmount := sdk.NewDecFromInt(amount).Quo(sourceChain.Redemptionratio).TruncateInt()
	if err := k.mintCoins(ctx, caller, sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, derivativeAmount)}); err != nil {
		return err
	}

	record.DelegationCoin = record.DelegationCoin.AddAmount(amount)

	k.SetDelegationRecord(ctx, recordID, record)

	return nil
}

// ProcessDelegationRecord start liquid stake on source chain with provide delegation records.
// This process will continue to advance the status of the DelegationRecord according to the IBC ack.
// So here just start and restart the process.
func (k *Keeper) ProcessDelegationRecord(ctx sdk.Context, curEpochNumber uint64, records []types.DelegationRecord) {
	for _, r := range records {
		// wait current done
		if curEpochNumber <= r.EpochNumber {
			continue
		}

		switch r.Status {
		case types.DelegationPending:
			k.handlePendingDelegationRecord(ctx, r)
		case types.DelegationTransferFailed:
			// become pending, transfer next epoch
		case types.DelegateFailed:
			// become transferred, delegate next epoch
		default:
			// do nothing
		}
	}
}

func (k Keeper) handlePendingDelegationRecord(ctx sdk.Context, record types.DelegationRecord) error {
	if record.DelegationCoin.Amount.IsZero() {
		return nil
	}

	transferCoin := record.DelegationCoin
	if !record.TransferredAmount.IsZero() {
		transferCoin = transferCoin.SubAmount(record.TransferredAmount)
	}

	if transferCoin.IsZero() && !record.DelegationCoin.IsZero() {
		// Only in the Delegate Epoch, no user participates in the Delegate,
		// but `Reinvest` has withdrawn rewards on the source chain
		return k.AfterDelegateTransfer(ctx, &record, true)
	}

	sourceChain, _ := k.GetSourceChain(ctx, record.ChainID)

	// send token from sourceChain's DelegateAddress to sourceChain's UnboudAddress
	if err := k.sendCoinsFromAccountToAccount(ctx,
		sdk.MustAccAddressFromBech32(sourceChain.EcsrowAddress),
		sdk.MustAccAddressFromBech32(sourceChain.DelegateAddress),
		sdk.Coins{record.DelegationCoin},
	); err != nil {
		return err
	}

	portID, err := icatypes.NewControllerPortID(sourceChain.DelegateAddress)
	if err != nil {
		return err
	}
	hostAddr, found := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, portID)
	if !found {
		return sdkerrors.Wrapf(types.ErrICANotFound, "address %s", sourceChain.DelegateAddress)
	}

	// TODO timeout ?
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	msg := ibctransfertypes.MsgTransfer{
		SourcePort:       ibctransfertypes.PortID,
		SourceChannel:    sourceChain.TransferChannelID,
		Token:            record.DelegationCoin,
		Sender:           sourceChain.DelegateAddress,
		Receiver:         hostAddr,
		TimeoutHeight:    ibcclienttypes.Height{},
		TimeoutTimestamp: uint64(timeoutTimestamp),
		Memo:             "",
	}

	resp, err := k.ibcTransferKeeper.Transfer(ctx, &msg)
	if err != nil {
		return err
	}

	bzArg := sdk.Uint64ToBigEndian(record.Id)
	callback := types.IBCCallback{
		CallType: types.DelegateTransferCall,
		Args:     string(bzArg),
	}

	// save ibc callback, wait ibc ack
	k.SetCallBack(ctx, msg.SourceChannel, msg.SourcePort, resp.Sequence, &callback)

	// update & save record
	record.Status = types.DelegationTransferring
	k.SetDelegationRecord(ctx, record.Id, &record)

	return nil
}

func (k Keeper) AfterDelegateTransfer(ctx sdk.Context, record *types.DelegationRecord, successfulTransfer bool) error {
	if !successfulTransfer {
		record.Status = types.DelegationTransferFailed
		k.SetDelegationRecord(ctx, record.Id, record)
		return nil
	}

	sourceChain, found := k.GetSourceChain(ctx, record.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", record.ChainID)
	}

	portID, err := icatypes.NewControllerPortID(sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	sourceChainDelegateAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, portID)
	sourceChainDelegateAddress := sdk.MustAccAddressFromBech32(sourceChainDelegateAddr)

	allocatedFunds := sourceChain.AllocateFundsForValidator(record.DelegationCoin.Amount)

	stakingMsgs := make([]proto.Message, 0)

	for _, valFund := range allocatedFunds {
		valAddress, err := sdk.ValAddressFromBech32(valFund.Address)
		if err != nil {
			return err
		}

		stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
			sourceChainDelegateAddress,
			valAddress,
			sdk.NewCoin(sourceChain.NativeDenom, valFund.Amount),
		))
	}

	sequence, portID, err := k.sendIBCMsg(ctx, stakingMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	record.Status = types.Delegating
	k.SetDelegationRecord(ctx, record.Id, record)

	bzArg := sdk.Uint64ToBigEndian(record.Id)
	callback := types.IBCCallback{
		CallType: types.DelegateCall,
		Args:     string(bzArg),
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	// save ibc callback, wait ibc ack
	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) AfterCrosschainDelegate(ctx sdk.Context, record *types.DelegationRecord, delegationSuccessful bool) error {
	if !delegationSuccessful {
		record.Status = types.DelegateFailed
		k.SetDelegationRecord(ctx, record.Id, record)
		return nil
	}

	sourceChain, found := k.GetSourceChain(ctx, record.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", record.ChainID)
	}

	record.Status = types.DelegationDone

	k.SetDelegationRecord(ctx, record.Id, record)

	sourceChain.UpdateWithDelegationRecord(record)

	// save source chain
	k.SetSourceChain(ctx, sourceChain)

	return nil
}

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
func (k *Keeper) Delegate(ctx sdk.Context, chainID string, amount math.Int, delegator sdk.AccAddress) error {
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

	sourceChainDelegatorAccAddress := sdk.MustAccAddressFromBech32(sourceChain.DelegateAddress)
	// transfer ibc token to sourcechain's delegation account
	if err := k.sendCoinsFromAccountToAccount(ctx, delegator, sourceChainDelegatorAccAddress, sdk.Coins{sdk.NewCoin(sourceChain.IbcDenom, amount)}); err != nil {
		return err
	}

	// TODO TruncateInt calculations can be huge precision error
	derivativeCoinAmount := amount.Mul(sourceChain.Redemptionratio.TruncateInt())
	if err := k.mintCoins(ctx, delegator, sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, derivativeCoinAmount)}); err != nil {
		return err
	}

	record.DelegationCoin = record.DelegationCoin.AddAmount(amount)

	k.SetDelegationRecord(ctx, recordID, record)

	return nil
}

// BeginLiquidStake start liquid stake on source chain with provide delegation records.
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
	// TODO The errors must be hanndler
	sourceChain, _ := k.GetSourceChain(ctx, record.ChainID)
	portID, _ := icatypes.NewControllerPortID(sourceChain.DelegateAddress)
	hostAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, portID)

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
	// TODO, sort map
	for valAddr, amount := range allocatedFunds {
		valAddress, err := sdk.ValAddressFromBech32(valAddr)
		if err != nil {
			return err
		}

		stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
			sourceChainDelegateAddress,
			valAddress,
			sdk.NewCoin(sourceChain.NativeDenom, amount),
		))
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, stakingMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// TODO timeout ?
	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	sequence, err := k.icaCtlKeeper.SendTx(ctx, nil, sourceChain.ConnectionID, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
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

	sourceChain.UpdateWithDelegationRecord(record)

	// save source chain
	k.SetSourceChain(ctx, sourceChain)

	return nil
}

// func (k Keeper) handleTransferFailedDelegationRecord(ctx sdk.Context, record types.DelegationRecord) {
// }

// func (k Keeper) handleDelegateFailedDelegationRecord(ctx sdk.Context, record types.DelegationRecord) {
// }

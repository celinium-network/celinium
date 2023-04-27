package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

// Delegate performs a liquid stake delegation. delegator transfer the ibcToken to module account then
// get derivative token by the rate.
func (k *Keeper) Delegate(ctx sdk.Context, chainID string, amount math.Int, caller sdk.AccAddress) (*types.ProxyDelegation, error) {
	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", chainID)
	}

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", appparams.DelegationEpochIdentifier)
	}

	currentEpoch := uint64(epochInfo.CurrentEpoch)
	delegationID, found := k.GetChianProxyDelegationID(ctx, chainID, currentEpoch)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoExistProxyDelegation, "chainID %s, epoch %d", chainID, currentEpoch)
	}

	proxyDelegation, found := k.GetProxyDelegation(ctx, delegationID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoExistProxyDelegation, "chainID %s, epoch %d, recorID %d", chainID, currentEpoch, delegationID)
	}

	ecsrowAccAddress := sdk.MustAccAddressFromBech32(sourceChain.EcsrowAddress)
	// transfer ibc token to sourcechain's ecsrow account
	if err := k.sendCoinsFromAccountToAccount(ctx, caller,
		ecsrowAccAddress, sdk.Coins{sdk.NewCoin(sourceChain.IbcDenom, amount)}); err != nil {
		return nil, err
	}

	derivativeAmount := sdk.NewDecFromInt(amount).Quo(sourceChain.Redemptionratio).TruncateInt()
	if err := k.mintCoins(ctx, caller, sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, derivativeAmount)}); err != nil {
		return nil, err
	}

	proxyDelegation.Coin = proxyDelegation.Coin.AddAmount(amount)

	k.SetProxyDelegation(ctx, delegationID, proxyDelegation)

	return proxyDelegation, nil
}

// ProcessProxyDelegation start liquid stake on source chain with provide delegation records.
// This process will continue to advance the status of the ProxyDelegation according to the IBC ack.
// So here just start and restart the process.
func (k *Keeper) ProcessProxyDelegation(ctx sdk.Context, curEpochNumber uint64, records []types.ProxyDelegation) {
	for _, r := range records {
		// wait current done
		if curEpochNumber <= r.EpochNumber {
			continue
		}

		switch r.Status {
		case types.ProxyDelegationPending:
			k.handlePendingProxyDelegation(ctx, r)
		case types.ProxyDelegationTransferFailed:
			// become pending, transfer next epoch
		case types.ProxyDelegationFailed:
			// become transferred, delegate next epoch
		default:
			// do nothing
		}
	}
}

func (k Keeper) handlePendingProxyDelegation(ctx sdk.Context, record types.ProxyDelegation) error {
	if record.Coin.Amount.IsZero() {
		return nil
	}

	transferCoin := record.Coin
	if !record.TransferredAmount.IsZero() {
		transferCoin = transferCoin.SubAmount(record.TransferredAmount)
	}

	if transferCoin.IsZero() && !record.Coin.IsZero() {
		// Only in the Delegate Epoch, no user participates in the Delegate,
		// but `Reinvest` has withdrawn rewards on the source chain
		return k.AfterProxyDelegationTransfer(ctx, &record, true)
	}

	sourceChain, _ := k.GetSourceChain(ctx, record.ChainID)

	// send token from sourceChain's DelegateAddress to sourceChain's UnboudAddress
	if err := k.sendCoinsFromAccountToAccount(ctx,
		sdk.MustAccAddressFromBech32(sourceChain.EcsrowAddress),
		sdk.MustAccAddressFromBech32(sourceChain.DelegateAddress),
		sdk.Coins{record.Coin},
	); err != nil {
		return err
	}

	hostAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	timeoutTimestamp := ctx.BlockTime().UnixNano() + types.DefaultIBCTransferTimeoutNanos
	msg := ibctransfertypes.MsgTransfer{
		SourcePort:       ibctransfertypes.PortID,
		SourceChannel:    sourceChain.TransferChannelID,
		Token:            record.Coin,
		Sender:           sourceChain.DelegateAddress,
		Receiver:         hostAddr,
		TimeoutHeight:    ibcclienttypes.Height{},
		TimeoutTimestamp: uint64(timeoutTimestamp),
		Memo:             "",
	}

	ctx.Logger().Info(fmt.Sprintf("transfer pending delegation record epoch: %d coin: %v",
		record.EpochNumber, record.Coin))

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
	record.Status = types.ProxyDelegationTransferring
	k.SetProxyDelegation(ctx, record.Id, &record)

	return nil
}

func (k Keeper) AfterProxyDelegationTransfer(ctx sdk.Context, record *types.ProxyDelegation, successfulTransfer bool) error {
	if !successfulTransfer {
		record.Status = types.ProxyDelegationTransferFailed
		k.SetProxyDelegation(ctx, record.Id, record)
		return nil
	}

	sourceChain, found := k.GetSourceChain(ctx, record.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", record.ChainID)
	}

	sourceChainDelegateAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}
	allocTokenVals := sourceChain.AllocateTokenForValidator(record.Coin.Amount)

	stakingMsgs := make([]proto.Message, 0)
	for _, val := range allocTokenVals.Validators {
		stakingMsgs = append(stakingMsgs, &stakingtypes.MsgDelegate{
			DelegatorAddress: sourceChainDelegateAddr,
			ValidatorAddress: val.Address,
			Amount: sdk.Coin{
				Denom:  sourceChain.NativeDenom,
				Amount: val.TokenAmount,
			},
		})
	}

	sequence, portID, err := k.sendIBCMsg(ctx, stakingMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	record.Status = types.ProxyDelegating
	k.SetProxyDelegation(ctx, record.Id, record)

	callbackArgs := types.DelegateCallbackArgs{
		Validators:        allocTokenVals.Validators,
		ProxyDelegationID: record.Id,
	}

	bzArg := k.cdc.MustMarshal(&callbackArgs)

	callback := types.IBCCallback{
		CallType: types.DelegateCall,
		Args:     string(bzArg),
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	// save ibc callback, wait ibc ack
	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) AfterProxyDelegationDone(ctx sdk.Context, delegateCallbackArgs *types.DelegateCallbackArgs, delegationSuccessful bool) error {
	delegation, found := k.GetProxyDelegation(ctx, delegateCallbackArgs.ProxyDelegationID)
	if !found {
		return nil
	}

	k.Logger(ctx).Info(fmt.Sprintf("delegateCallbackHandler, chainID %s epoch %d", delegation.ChainID, delegation.EpochNumber))

	if !delegationSuccessful {
		delegation.Status = types.ProxyDelegationFailed
		k.SetProxyDelegation(ctx, delegation.Id, delegation)
		return nil
	}

	sourceChain, found := k.GetSourceChain(ctx, delegation.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", delegation.ChainID)
	}

	delegation.Status = types.ProxyDelegationDone

	k.SetProxyDelegation(ctx, delegation.Id, delegation)

	sourceChain.UpdateWithDelegatedValidators(delegateCallbackArgs.Validators)

	k.SetSourceChain(ctx, sourceChain)

	return nil
}

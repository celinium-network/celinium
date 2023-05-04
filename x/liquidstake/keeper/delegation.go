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
	for i := 0; i < len(records); i++ {
		if curEpochNumber <= records[i].EpochNumber {
			continue
		}

		switch records[i].Status {
		case types.ProxyDelegationPending:
			k.handlePendingProxyDelegation(ctx, records[i])
		case types.ProxyDelegationFailed:
			// become transferred, retry delegate next epoch
			records[i].Status = types.ProxyDelegationTransferred
			k.SetProxyDelegation(ctx, curEpochNumber, &records[i])
		case types.ProxyDelegationTransferred:
			k.afterProxyDelegationTransfer(ctx, &records[i])
		default:
			// do nothing
		}
	}
}

func (k Keeper) handlePendingProxyDelegation(ctx sdk.Context, delegation types.ProxyDelegation) error {
	if delegation.Coin.Amount.IsZero() {
		return nil
	}

	transferCoin := delegation.Coin
	if !delegation.ReinvestAmount.IsZero() {
		transferCoin = transferCoin.SubAmount(delegation.ReinvestAmount)
	}

	if transferCoin.IsZero() && !delegation.Coin.IsZero() {
		// Only in the Delegate Epoch, no user participates in the Delegate,
		// but `Reinvest` has withdrawn rewards on the source chain
		return k.afterProxyDelegationTransfer(ctx, &delegation)
	}

	sourceChain, _ := k.GetSourceChain(ctx, delegation.ChainID)

	// send token from sourceChain's DelegateAddress to sourceChain's UnboudAddress
	if err := k.sendCoinsFromAccountToAccount(ctx,
		sdk.MustAccAddressFromBech32(sourceChain.EcsrowAddress),
		sdk.MustAccAddressFromBech32(sourceChain.DelegateAddress),
		sdk.Coins{delegation.Coin},
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
		Token:            delegation.Coin,
		Sender:           sourceChain.DelegateAddress,
		Receiver:         hostAddr,
		TimeoutHeight:    ibcclienttypes.Height{},
		TimeoutTimestamp: uint64(timeoutTimestamp),
		Memo:             "",
	}

	ctx.Logger().Info(fmt.Sprintf("transfer pending delegation record epoch: %d coin: %v",
		delegation.EpochNumber, delegation.Coin))

	resp, err := k.ibcTransferKeeper.Transfer(ctx, &msg)
	if err != nil {
		return err
	}

	bzArg := sdk.Uint64ToBigEndian(delegation.Id)
	callback := types.IBCCallback{
		CallType: types.DelegateTransferCall,
		Args:     string(bzArg),
	}

	// save ibc callback, wait ibc ack
	k.SetCallBack(ctx, msg.SourceChannel, msg.SourcePort, resp.Sequence, &callback)

	// update & save record
	delegation.Status = types.ProxyDelegationTransferring
	k.SetProxyDelegation(ctx, delegation.Id, &delegation)

	return nil
}

func (k Keeper) afterProxyDelegationTransfer(ctx sdk.Context, delegation *types.ProxyDelegation) error {
	sourceChain, found := k.GetSourceChain(ctx, delegation.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", delegation.ChainID)
	}

	sourceChainDelegateAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	allocTokenVals := sourceChain.AllocateTokenForValidator(delegation.Coin.Amount)

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

	delegation.Status = types.ProxyDelegating
	k.SetProxyDelegation(ctx, delegation.Id, delegation)

	callbackArgs := types.DelegateCallbackArgs{
		Validators:        allocTokenVals.Validators,
		ProxyDelegationID: delegation.Id,
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

func (k Keeper) afterProxyDelegationDone(ctx sdk.Context, delegateCallbackArgs *types.DelegateCallbackArgs, delegationSuccessful bool) error {
	delegation, found := k.GetProxyDelegation(ctx, delegateCallbackArgs.ProxyDelegationID)
	if !found {
		return types.ErrNoExistProxyDelegation
	}

	k.Logger(ctx).Info(fmt.Sprintf("delegateCallbackHandler, chainID %s epoch %d", delegation.ChainID, delegation.EpochNumber))

	if !delegationSuccessful {
		delegation.Status = types.ProxyDelegationFailed
		k.SetProxyDelegation(ctx, delegation.Id, delegation)
		return types.ErrCallbackMismatch
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

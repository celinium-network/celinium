package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

type callbackHandler func(*Keeper, sdk.Context, *types.IBCCallback, []byte) error

var callbackHandlerRegistry map[types.CallType]callbackHandler

func init() {
	callbackHandlerRegistry = make(map[types.CallType]callbackHandler)

	callbackHandlerRegistry[types.DelegateTransferCall] = delegateTransferCallbackHandler
	callbackHandlerRegistry[types.DelegateCall] = delegateCallbackHandler
	callbackHandlerRegistry[types.UndelegateCall] = undelegateCallbackHandler
	callbackHandlerRegistry[types.WithdrawUnbondCall] = withdrawUnbondCallbackHandler
	callbackHandlerRegistry[types.WithdrawDelegateRewardCall] = withdrawDelegateRewardCallbackHandler
	callbackHandlerRegistry[types.TransferRewardCall] = transferRewardCallbackHandler
	callbackHandlerRegistry[types.SetWithdrawAddressCall] = setWithdrawAddressCallbackHandler
}

func delegateTransferCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	proxyDelegationID := sdk.BigEndianToUint64([]byte(callback.Args))
	delegation, found := k.GetProxyDelegation(ctx, proxyDelegationID)
	if !found {
		return types.ErrNoExistProxyDelegation
	}
	k.Logger(ctx).Info(fmt.Sprintf("ibc callback `DelegateTransferCall`, chainID %s epoch %d",
		delegation.ChainID, delegation.EpochNumber))

	if delegation.Status != types.ProxyDelegationTransferring {
		k.Logger(ctx).Error(fmt.Sprintf("ibc callback `DelegateTransferCall` with wrong status chainID %s epoch %d",
			delegation.ChainID, delegation.EpochNumber))
		return types.ErrCallbackMismatch
	}

	if _, err := GetResultFromAcknowledgement(acknowledgement); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("proxydelegation ibc transfer failed. chainID %s, epoch %d",
			delegation.ChainID, delegation.EpochNumber))

		// let delegation become pending, try ibc send in next epoch.
		delegation.Status = types.ProxyDelegationPending
		k.SetProxyDelegation(ctx, delegation.Id, delegation)

		return err
	}

	return k.afterProxyDelegationTransfer(ctx, delegation)
}

func delegateCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) (err error) {
	var delegateCallbackArgs types.DelegateCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &delegateCallbackArgs)

	checkProxyDelegationAck := true

	defer func() { err = k.afterProxyDelegationDone(ctx, &delegateCallbackArgs, checkProxyDelegationAck) }()

	ackRes, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		checkProxyDelegationAck = false
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(ackRes, &txMsgData); err != nil {
		checkProxyDelegationAck = false
		return err
	}

	respLen := 0
	for _, r := range txMsgData.MsgResponses {
		if !strings.Contains(r.TypeUrl, "MsgDelegateResponse") { // "/cosmos.staking.v1beta1.MsgDelegateResponse"
			continue
		}
		response := stakingtypes.MsgDelegate{}
		if err := k.cdc.Unmarshal(r.Value, &response); err != nil {
			checkProxyDelegationAck = false
			return err
		}
		respLen++
	}

	if checkProxyDelegationAck = (respLen == len(delegateCallbackArgs.Validators)); !checkProxyDelegationAck {
		return types.ErrCallbackMismatch
	}
	return nil
}

func undelegateCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	var completeTime time.Time
	for _, r := range txMsgData.MsgResponses {
		if strings.Contains(r.TypeUrl, "MsgUndelegateResponse") {
			response := stakingtypes.MsgUndelegateResponse{}
			if err := k.cdc.Unmarshal(r.Value, &response); err != nil {
				return nil
			}
			completeTime = response.CompletionTime
		}
	}

	var unbondCallArgs types.UnbondCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &unbondCallArgs)

	epochUnbondings, found := k.GetEpochProxyUnboundings(ctx, unbondCallArgs.Epoch)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown epochUnbonding, chainID: %s, epoch %d",
			unbondCallArgs.ChainID, unbondCallArgs.Epoch)
	}

	for i := 0; i < len(epochUnbondings.Unbondings); i++ {
		if epochUnbondings.Unbondings[i].ChainID != unbondCallArgs.ChainID {
			continue
		}
		epochUnbondings.Unbondings[i].UnbondTime = uint64(completeTime.UnixNano())
		epochUnbondings.Unbondings[i].Status = types.ProxyUnbondingWaitting

		// update sourcechain
		sourceChain, found := k.GetSourceChain(ctx, unbondCallArgs.ChainID)
		if !found {
			return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", unbondCallArgs.ChainID)
		}

		sourceChain.UpdateWithUnbondingValidators(unbondCallArgs.Validators)

		burnedCoin := sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, epochUnbondings.Unbondings[i].BurnedDerivativeAmount)}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnedCoin); err != nil {
			return err
		}
		k.SetSourceChain(ctx, sourceChain)
	}

	k.SetEpochProxyUnboundings(ctx, epochUnbondings)

	return nil
}

func withdrawUnbondCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	var unbondCallArgs types.UnbondCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &unbondCallArgs)
	epochUnbondings, found := k.GetEpochProxyUnboundings(ctx, unbondCallArgs.Epoch)
	if !found {
		return nil
	}

	unbondings := epochUnbondings.Unbondings
	unbondingLen := len(unbondings)
	for i := 0; i < unbondingLen; i++ {
		if unbondings[i].ChainID != unbondCallArgs.ChainID {
			continue
		}
		unbondings[i].Status = types.ProxyUnbondingDone

		for _, userUnDelegationID := range unbondings[i].UserUnbondingIds {
			userUnbonding, found := k.GetUserUnbondingID(ctx, userUnDelegationID)
			if !found {
				continue
			}
			userUnbonding.CliamStatus = types.UserUnbondingClaimable
			k.SetUserUnbonding(ctx, userUnbonding)
		}
	}
	k.SetEpochProxyUnboundings(ctx, epochUnbondings)
	return nil
}

func withdrawDelegateRewardCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	var callbackArgs types.WithdrawDelegateRewardCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &callbackArgs)
	totalReward := math.ZeroInt()
	sourceChain, found := k.GetSourceChain(ctx, callbackArgs.ChainID)
	if !found {
		return nil
	}
	for _, r := range txMsgData.MsgResponses {
		if strings.Contains(r.TypeUrl, "MsgWithdrawDelegatorRewardResponse") {
			response := distrtypes.MsgWithdrawDelegatorRewardResponse{}
			if err := k.cdc.Unmarshal(r.Value, &response); err != nil {
				return nil
			}
			for _, c := range response.Amount {
				if c.Amount.IsNil() || c.Amount.IsZero() {
					continue
				}
				totalReward = totalReward.Add(c.Amount)
			}
		}
	}
	if !totalReward.IsZero() {
		k.AfterWithdrawDelegateReward(ctx, sourceChain, totalReward)
	}

	return nil
}

func transferRewardCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	var callbackArgs types.TransferRewardCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &callbackArgs)
	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	if !found {
		return nil
	}

	currentEpoch := uint64(epochInfo.CurrentEpoch)
	delegationID, found := k.GetChianProxyDelegationID(ctx, callbackArgs.ChainID, currentEpoch)
	if !found {
		return nil
	}

	delegation, found := k.GetProxyDelegation(ctx, delegationID)
	if !found {
		return nil
	}

	delegation.Coin = delegation.Coin.AddAmount(callbackArgs.Amount)
	delegation.ReinvestAmount = delegation.ReinvestAmount.Add(callbackArgs.Amount)
	k.SetProxyDelegation(ctx, delegationID, delegation)
	return nil
}

func setWithdrawAddressCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, acknowledgement []byte) error {
	return nil
}

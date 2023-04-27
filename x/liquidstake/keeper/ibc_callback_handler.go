package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

type callbackHandler func(*Keeper, sdk.Context, *types.IBCCallback, []*codectypes.Any) error

var callbackHandlerRegistry map[types.CallType]callbackHandler

func init() {
	callbackHandlerRegistry = make(map[types.CallType]callbackHandler)

	callbackHandlerRegistry[types.DelegateTransferCall] = delegateTransferCallbackHandler
	callbackHandlerRegistry[types.DelegateCall] = delegateCallbackHandler
	callbackHandlerRegistry[types.UnbondCall] = unbondCallbackHandler
	callbackHandlerRegistry[types.WithdrawUnbondCall] = withdrawUnbondCallbackHandler
	callbackHandlerRegistry[types.WithdrawDelegateRewardCall] = withdrawDelegateRewardCallbackHandler
	callbackHandlerRegistry[types.TransferRewardCall] = transferRewardCallbackHandler
	callbackHandlerRegistry[types.SetWithdrawAddressCall] = setWithdrawAddressCallbackHandler
}

func delegateTransferCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, response []*codectypes.Any) error {
	proxyDelegationID := sdk.BigEndianToUint64([]byte(callback.Args))
	record, found := k.GetProxyDelegation(ctx, proxyDelegationID)
	if !found {
		return nil
	}

	k.Logger(ctx).Info(fmt.Sprintf("delegateTransferCallbackHandler, chainID %s epoch %d", record.ChainID, record.EpochNumber))
	k.AfterProxyDelegationTransfer(ctx, record, true)
	return nil
}

func delegateCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, response []*codectypes.Any) error {
	proxyDelegationID := sdk.BigEndianToUint64([]byte(callback.Args))
	record, found := k.GetProxyDelegation(ctx, proxyDelegationID)
	if !found {
		return nil
	}

	k.Logger(ctx).Info(fmt.Sprintf("delegateCallbackHandler, chainID %s epoch %d", record.ChainID, record.EpochNumber))
	k.AfterProxyDelegationDone(ctx, record, true)

	return nil
}

func unbondCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any) error {
	var completeTime time.Time
	for _, r := range responses {
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

	epochUnbondings, found := k.GetEpochUnboundings(ctx, unbondCallArgs.Epoch)
	if !found {
		return nil
	}

	for i := 0; i < len(epochUnbondings.Unbondings); i++ {
		if epochUnbondings.Unbondings[i].ChainID != unbondCallArgs.ChainID {
			continue
		}
		epochUnbondings.Unbondings[i].UnbondTIme = uint64(completeTime.UnixNano())
		epochUnbondings.Unbondings[i].Status = types.ProxyUnbondingWaitting

		// update sourcechain
		sourceChain, found := k.GetSourceChain(ctx, unbondCallArgs.ChainID)
		if !found {
			return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", unbondCallArgs.ChainID)
		}
		sourceChain.StakedAmount = sourceChain.StakedAmount.Sub(epochUnbondings.Unbondings[i].RedeemNativeToken.Amount)

		burnedCoin := sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, epochUnbondings.Unbondings[i].BurnedDerivativeAmount)}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnedCoin); err != nil {
			return err
		}
		k.SetSourceChain(ctx, sourceChain)
	}

	// save
	k.SetEpochUnboundings(ctx, epochUnbondings)

	return nil
}

func withdrawUnbondCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any) error {
	var unbondCallArgs types.UnbondCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &unbondCallArgs)
	epochUnbondings, found := k.GetEpochUnboundings(ctx, unbondCallArgs.Epoch)
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

		for _, userUnDelegationID := range unbondings[i].UserUnbondRecordIds {
			userUndelegation, found := k.GetUndelegationRecordByID(ctx, userUnDelegationID)
			if !found {
				continue
			}
			userUndelegation.CliamStatus = types.UserUnbondingClaimable
			k.SetUndelegationRecord(ctx, userUndelegation)
		}
	}
	k.SetEpochUnboundings(ctx, epochUnbondings)
	return nil
}

func withdrawDelegateRewardCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any) error {
	var callbackArgs types.WithdrawDelegateRewardCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &callbackArgs)
	totalReward := math.ZeroInt()
	sourceChain, found := k.GetSourceChain(ctx, callbackArgs.ChainID)
	if !found {
		return nil
	}
	for _, r := range responses {
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

func transferRewardCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any) error {
	var callbackArgs types.TransferRewardCallbackArgs
	k.cdc.MustUnmarshal([]byte(callback.Args), &callbackArgs)
	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	if !found {
		return nil
	}

	currentEpoch := uint64(epochInfo.CurrentEpoch)
	recordID, found := k.GetChianProxyDelegationID(ctx, callbackArgs.ChainID, currentEpoch)
	if !found {
		return nil
	}

	record, found := k.GetProxyDelegation(ctx, recordID)
	if !found {
		return nil
	}

	record.Coin = record.Coin.AddAmount(callbackArgs.Amount)
	record.TransferredAmount = record.TransferredAmount.Add(callbackArgs.Amount)
	k.SetProxyDelegation(ctx, recordID, record)
	return nil
}

func setWithdrawAddressCallbackHandler(k *Keeper, ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any) error {
	return nil
}

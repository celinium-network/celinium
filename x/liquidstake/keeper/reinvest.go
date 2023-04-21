package keeper

import (
	"cosmossdk.io/math"
	"github.com/gogo/protobuf/proto"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (k Keeper) Reinvest(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	for ; iterator.Valid(); iterator.Next() {
		sourcechain := &types.SourceChain{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		k.WithdrawDelegateReward(ctx, sourcechain)
	}
}

func (k Keeper) WithdrawDelegateReward(ctx sdk.Context, sourceChain *types.SourceChain) error {
	delegateAccAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	sendMsgs := make([]proto.Message, 0)

	for _, v := range sourceChain.Validators {
		sendMsgs = append(sendMsgs, &distrtypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegateAccAddr,
			ValidatorAddress: v.Address,
		})
	}

	sequence, portID, err := k.sendIBCMsg(ctx, sendMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	// TODO record length of sendmsgs?
	callbackArgs := types.WithdrawDelegateRewardCallbackArgs{
		ChainID: sourceChain.ChainID,
	}

	callbackArgsBz := k.cdc.MustMarshal(&callbackArgs)

	callback := types.IBCCallback{
		CallType: types.WithdrawDelegateRewardCall,
		Args:     string(callbackArgsBz),
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) AfterWithdrawDelegateReward(ctx sdk.Context, sourceChain *types.SourceChain, reward math.Int) error {
	delegateAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	rewardAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
	if err != nil {
		return err
	}

	sendMsgs := make([]proto.Message, 0)

	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: rewardAddr,
		ToAddress:   delegateAddr,
		Amount:      []sdk.Coin{sdk.NewCoin(sourceChain.NativeDenom, reward)},
	})

	sequence, portID, err := k.sendIBCMsg(ctx, sendMsgs, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
	if err != nil {
		return err
	}

	// TODO record length of sendmsgs?
	callbackArgs := types.TransferRewardCallbackArgs{
		ChainID: sourceChain.ChainID,
		Amount:  reward,
	}

	callbackArgsBz := k.cdc.MustMarshal(&callbackArgs)

	callback := types.IBCCallback{
		CallType: types.TransferRewardCall,
		Args:     string(callbackArgsBz),
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

// SetDistriWithdrawAddress set the sourcechain staking reward recipient.
// Only after successful, the sourcechain is available.
func (k Keeper) SetDistriWithdrawAddress(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	for ; iterator.Valid(); iterator.Next() {
		sourceChain := &types.SourceChain{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourceChain)

		delegateAccAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
		if err != nil {
			return err
		}

		rewardAccAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
		if err != nil {
			return err
		}

		sendMsgs := make([]proto.Message, 0)

		sendMsgs = append(sendMsgs, &distrtypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegateAccAddr,
			WithdrawAddress:  rewardAccAddr,
		})

		sequence, portID, err := k.sendIBCMsg(ctx, sendMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
		if err != nil {
			return err
		}

		callbackArgs := types.SetWithdrawMessageArgs{
			ChainID: sourceChain.ChainID,
		}

		callbackArgsBz := k.cdc.MustMarshal(&callbackArgs)

		callback := types.IBCCallback{
			CallType: types.SetWithdrawAddressCall,
			Args:     string(callbackArgsBz),
		}

		sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

		k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)
	}

	return nil
}

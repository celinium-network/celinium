package keeper

import (
	"time"

	"celinium/x/inter-staking/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	"github.com/gogo/protobuf/proto"
)

func (k Keeper) UnDelegate(ctx sdk.Context, chainID string, undelegationAmount sdk.Coin, delegator string) error {
	sourceChainMetadata, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}
	// check wheather the ica of source chain in control endpoint is available.
	if !k.SourceChainAvaiable(ctx, sourceChainMetadata.IbcConnectionId, sourceChainMetadata.ICAControlAddr) {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	// Check wheather the coin is the native token of the source chain.
	// if strings.Compare(undelegationAmount.Denom, sourceChainMetadata.SourceChainTraceDenom) != 0 {
	// 	return sdkerrors.Wrapf(types.ErrMismatchSourceCoin, "chainID: %s, expected: %s, get:",
	// 		chainID, sourceChainMetadata.SourceChainTraceDenom, undelegationAmount.Denom)
	// }

	delegationAmount := k.GetDelegation(ctx, delegator, chainID)
	if delegationAmount.Amount.LT(undelegationAmount.Amount) {
		return sdkerrors.Wrapf(types.ErrInsufficientDelegation, "exist: %s, expected %s", delegationAmount, undelegationAmount)
	}

	remindingAmount := delegationAmount.Sub(undelegationAmount)
	// remove key if remindingAmount == 0?
	k.SetDelegation(&ctx, delegator, chainID, remindingAmount)

	stragegyLen := len(sourceChainMetadata.DelegateStrategy)

	undelegateMsgs := make([]proto.Message, 0)
	usedAmount := sdkmath.NewInt(0)

	portID, err := icatypes.NewControllerPortID(sourceChainMetadata.ICAControlAddr)
	if err != nil {
		return nil
	}
	hostAddr, ok := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, sourceChainMetadata.IbcConnectionId, portID)
	if !ok {
		return sdk.ErrEmptyHexAddress
	}

	hostAddress := sdk.MustAccAddressFromBech32(hostAddr)

	for i := 0; i < stragegyLen-2; i++ {
		percentage := sdkmath.NewIntFromUint64(uint64(sourceChainMetadata.DelegateStrategy[i].Percentage))
		stakingAmount := undelegationAmount.Amount.Mul(percentage).BigInt()
		stakingAmount.Div(stakingAmount, types.PercentageDenominator.BigInt())
		usedAmount.Add(sdkmath.NewIntFromBigInt(stakingAmount))

		valAddress, err := sdk.ValAddressFromBech32(sourceChainMetadata.DelegateStrategy[i].ValidatorAddress)
		if err != nil {
			return err
		}

		undelegateMsgs = append(undelegateMsgs, stakingtypes.NewMsgUndelegate(
			hostAddress,
			valAddress,
			sdk.NewCoin(sourceChainMetadata.SourceChainDenom, sdkmath.NewIntFromBigInt(stakingAmount)),
		))
	}

	if !usedAmount.Equal(delegationAmount.Amount) {
		valAddress, err := sdk.ValAddressFromBech32(sourceChainMetadata.DelegateStrategy[stragegyLen-1].ValidatorAddress)
		if err != nil {
			return err
		}
		undelegateMsgs = append(undelegateMsgs, stakingtypes.NewMsgUndelegate(
			hostAddress,
			valAddress,
			sdk.NewCoin(sourceChainMetadata.SourceChainDenom, delegationAmount.Amount.Sub(usedAmount)),
		))
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, undelegateMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	sequence, err := k.icaControllerKeeper.SendTx(ctx, nil, sourceChainMetadata.IbcConnectionId, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}

	// push into undelegation pending queue. I think a separate queue is needed ?
	k.PushDelegationTaskQueue(&ctx, types.PendingUndelegationQueueKey, sequence, &types.DelegationTask{
		ChainId:   chainID,
		Delegator: delegator,
		Amount:    undelegationAmount,
	})

	return nil
}

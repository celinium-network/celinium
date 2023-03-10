package keeper

import (
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"

	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"celinium/x/inter-staking/types"
)

func (k Keeper) Delegate(ctx sdk.Context, chainID string, coin sdk.Coin, delegator string) error {
	sourceChainMetadata, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	// check wheather the ica of source chain in control endpoint is available.
	if !k.SourceChainAvaiable(ctx, sourceChainMetadata.IbcConnectionId, sourceChainMetadata.ICAControlAddr) {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	// Check wheather the coin is the native token of the source chain.
	if strings.Compare(coin.Denom, sourceChainMetadata.StakingDenom) != 0 {
		return sdkerrors.Wrapf(types.ErrMismatchSourceCoin, "chainID: %s, expected: %s, get:",
			chainID, sourceChainMetadata.StakingDenom, coin.Denom)
	}

	if err := k.SendCoinsFromDelegatorToICA(ctx, delegator, sourceChainMetadata.ICAControlAddr, sdk.Coins{coin}); err != nil {
		return err
	}

	newDelegationTask := types.DelegationTask{
		ChainId:   chainID,
		Delegator: delegator,
		Amount:    coin,
	}

	k.PushDelegationTaskQueue(&ctx, types.PendingDelegationQueueKey, &newDelegationTask)

	return nil
}

func (k Keeper) SendCoinsFromDelegatorToICA(ctx sdk.Context, delegatorAddr string, icaCtlAddr string, coins sdk.Coins) error {
	delegatorAccount := sdk.MustAccAddressFromBech32(delegatorAddr)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAccount, types.ModuleName, coins); err != nil {
		return err
	}

	icaCtlAccount := sdk.MustAccAddressFromBech32(icaCtlAddr)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, icaCtlAccount, coins); err != nil {
		return err
	}

	return nil
}

func (k Keeper) ProcessDelegationTask(ctx sdk.Context) {
	k.ProcessPendingDelegationTask(ctx, 100)
}

func (k Keeper) ProcessPendingDelegationTask(ctx sdk.Context, maxTask int32) error {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingDelegationQueueKey)

	chainDelegation := make(map[string]sdk.Coin)
	userDelegations := make(map[string][]types.DelegationTask)

	// map is't order, keep the key sort in slice, then the map can be traversed in order.
	orderedChainIDKeys := make([]string, 0)

	// only get metadata for each source chain once.
	sourceChainMetadataCache := make(map[string]types.SourceChainMetadata)

	lastIndex := int32(0)
	for ; lastIndex < maxTask && iterator.Valid(); iterator.Next() {
		taskSlice := types.DelegationTasks{}
		types.MustUnMarshalProtoType(k.cdc, iterator.Value(), &taskSlice)

		for _, t := range taskSlice.DelegationTasks {
			chainCoin, ok := chainDelegation[t.ChainId]
			if !ok {
				chainDelegation[t.ChainId] = t.Amount
				orderedChainIDKeys = append(orderedChainIDKeys, t.ChainId)

				sourceChainMetadata, _ := k.GetSourceChain(ctx, t.ChainId)
				sourceChainMetadataCache[t.ChainId] = *sourceChainMetadata

			} else {
				chainDelegation[t.ChainId] = chainCoin.Add(t.Amount)
			}

			// delete task from pending queue
			store.Delete(iterator.Key())
			// store task to preparing queue

			ts := userDelegations[t.ChainId]
			ts = append(ts, t)
			userDelegations[t.ChainId] = ts

			lastIndex++
		}
	}

	for _, chainID := range orderedChainIDKeys {
		metadata := sourceChainMetadataCache[chainID]
		portID, _ := icatypes.NewControllerPortID(metadata.ICAControlAddr)

		stragegyLen := len(metadata.DelegateStrategy)
		stakingMsgs := make([]proto.Message, 0)

		totalCoin := chainDelegation[chainID]
		usedAmount := math.NewInt(0)
		for i := 0; i < stragegyLen-2; i++ {
			percentage := math.NewIntFromUint64(uint64(metadata.DelegateStrategy[i].Percentage))
			stakingAmount := totalCoin.Amount.Mul(percentage).BigInt()
			stakingAmount.Div(stakingAmount, types.PercentageDenominator.BigInt())
			usedAmount.Add(math.NewIntFromBigInt(stakingAmount))
			stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
				sdk.AccAddress(metadata.ICAControlAddr),
				sdk.ValAddress(metadata.DelegateStrategy[i].ValidatorAddress),
				sdk.NewCoin(totalCoin.Denom, math.NewIntFromBigInt(stakingAmount)),
			))
		}

		if !usedAmount.Equal(totalCoin.Amount) {
			stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
				sdk.AccAddress(metadata.ICAControlAddr),
				sdk.ValAddress(metadata.DelegateStrategy[stragegyLen-1].ValidatorAddress),
				sdk.NewCoin(totalCoin.Denom, totalCoin.Amount.Sub(usedAmount)),
			))
		}

		data, err := icatypes.SerializeCosmosTx(k.cdc, stakingMsgs)
		if err != nil {
			continue
		}

		packetData := icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: data,
		}

		timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
		sequence, err := k.icaControllerKeeper.SendTx(ctx, nil, metadata.IbcConnectionId, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
		if err != nil {
			continue
		}

		useTaskLen := len(userDelegations[chainID])

		for i := 0; i < useTaskLen; i++ {
			preparingDelegationTask := userDelegations[chainID][i]
			preparingDelegationTask.DoneSingal = sequence
			k.PushDelegationTaskQueue(&ctx, types.PreparingDelegationQueueKey, &preparingDelegationTask)
		}
	}

	return nil
}

func (k Keeper) HandleTransferAcknowledgementPacket(packet *channeltypes.Packet) {
	// advence delegation task from preparing to prepared
}

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
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
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
	if strings.Compare(coin.Denom, sourceChainMetadata.SourceChainTraceDenom) != 0 {
		return sdkerrors.Wrapf(types.ErrMismatchSourceCoin, "chainID: %s, expected: %s, get:",
			chainID, sourceChainMetadata.SourceChainTraceDenom, coin.Denom)
	}

	if err := k.SendCoinsFromDelegatorToICA(ctx, delegator, sourceChainMetadata.ICAControlAddr, sdk.Coins{coin}); err != nil {
		return err
	}

	newDelegationTask := types.DelegationTask{
		ChainId:   chainID,
		Delegator: delegator,
		Amount:    coin,
	}

	k.PushDelegationTaskQueue(&ctx, types.PendingDelegationQueueKey, uint64(ctx.BlockHeight()), &newDelegationTask)

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

		hostAddr, _ := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, metadata.IbcConnectionId, portID)
		hostAddress := sdk.MustAccAddressFromBech32(hostAddr)

		totalCoin := chainDelegation[chainID]

		// timeout should be task parameter or never timeout?
		timeoutHeight := clienttypes.NewHeight(0, 1000000)
		transferMsg := transfertypes.NewMsgTransfer(
			transfertypes.PortID,
			metadata.IbcTransferChannelId,
			sdk.NewCoin(metadata.SourceChainTraceDenom, totalCoin.Amount),
			metadata.ICAControlAddr,
			hostAddr,
			timeoutHeight,
			0,
			"",
		)

		k.ibcTransferKeeper.Transfer(ctx, transferMsg)

		stakingMsgs := make([]proto.Message, 0)

		usedAmount := math.NewInt(0)
		for i := 0; i < stragegyLen-2; i++ {
			percentage := math.NewIntFromUint64(uint64(metadata.DelegateStrategy[i].Percentage))
			stakingAmount := totalCoin.Amount.Mul(percentage).BigInt()
			stakingAmount.Div(stakingAmount, types.PercentageDenominator.BigInt())
			usedAmount.Add(math.NewIntFromBigInt(stakingAmount))

			valAddress, err := sdk.ValAddressFromBech32(metadata.DelegateStrategy[i].ValidatorAddress)
			if err != nil {
				return err
			}

			stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
				hostAddress,
				valAddress,
				sdk.NewCoin(metadata.SourceChainDenom, math.NewIntFromBigInt(stakingAmount)),
			))
		}

		if !usedAmount.Equal(totalCoin.Amount) {
			valAddress, err := sdk.ValAddressFromBech32(metadata.DelegateStrategy[stragegyLen-1].ValidatorAddress)
			if err != nil {
				return err
			}
			stakingMsgs = append(stakingMsgs, stakingtypes.NewMsgDelegate(
				hostAddress,
				valAddress,
				sdk.NewCoin(metadata.SourceChainDenom, totalCoin.Amount.Sub(usedAmount)),
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
			k.PushDelegationTaskQueue(&ctx, types.PreparingDelegationQueueKey, sequence, &preparingDelegationTask)
		}
	}

	return nil
}

func (k Keeper) OnAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet) {
	// remove delegation from preparing queue
	preparingDelegationTasks := k.GetDelegationQueueSlice(&ctx, types.PreparingDelegationQueueKey, packet.Sequence)
	if len(preparingDelegationTasks) == 0 {
		return
	}

	for _, task := range preparingDelegationTasks {
		k.SetDelegationForDelegator(&ctx, task)
	}

	// remove from preparing delegation queue
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDelegateQueueKey(types.PreparingDelegationQueueKey, packet.Sequence))
}

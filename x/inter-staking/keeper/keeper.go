package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"

	"celinium/x/inter-staking/types"
)

// Keeper of the x/inter-staking store
type Keeper struct {
	storeKey  storetypes.StoreKey
	cdc       codec.Codec
	authority string

	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	icaControllerKeeper icacontrollerkeeper.Keeper
}

func (k Keeper) SetSourceChain(ctx sdk.Context, chainID string, sourceChainMetadata *types.SourceChainMetadata) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetSourceChainMetadataKey([]byte(chainID)),
		types.MustMarshalProtoType(k.cdc, sourceChainMetadata))
}

func (k Keeper) GetSourceChain(ctx sdk.Context, chainID string) (sourceChainMetadata *types.SourceChainMetadata, found bool) {
	found, bz := k.SourceChainExist(ctx, chainID)
	if !found {
		return nil, false
	}

	err := types.UnMarshalProtoType(k.cdc, bz, sourceChainMetadata)
	if err != nil {
		return nil, false
	}

	return sourceChainMetadata, true
}

func (k Keeper) SourceChainExist(ctx sdk.Context, chainID string) (bool, []byte) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetSourceChainMetadataKey([]byte(chainID)))
	if bz == nil {
		return false, nil
	}
	return true, bz
}

func (k Keeper) SetSourceChainDelegation(ctx sdk.Context, chainID string, delegation *types.SourceChainDelegation) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetSourceChainDelegationKey([]byte(chainID)),
		types.MustMarshalProtoType(k.cdc, delegation))
}

func (k Keeper) GetSourceChainDelegation(ctx sdk.Context, chainID string) (delegation *types.SourceChainDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetSourceChainMetadataKey([]byte(chainID)))
	if bz == nil {
		return delegation, false
	}

	err := types.UnMarshalProtoType(k.cdc, bz, delegation)
	if err != nil {
		return nil, false
	}

	return delegation, true
}

func (k Keeper) PushDelegationTaskQueue(ctx *sdk.Context, delegationTask *types.DelegationTask) {
	tasks := k.GetDelegationQueueSlice(ctx, uint64(ctx.BlockHeight()))

	height := uint64(ctx.BlockHeight())
	if len(tasks) == 0 {
		k.SetDelegationQueueSlice(ctx, []types.DelegationTask{*delegationTask}, height)
	} else {
		tasks = append(tasks, *delegationTask)
		k.SetDelegationQueueSlice(ctx, tasks, height)
	}
}

func (k Keeper) GetDelegationQueueSlice(ctx *sdk.Context, height uint64) []types.DelegationTask {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDelegateQueueKey(height))

	if bz == nil {
		return []types.DelegationTask{}
	}

	tasks := types.DelegationTasks{}
	k.cdc.MustUnmarshal(bz, &tasks)

	return tasks.DelegationTasks
}

func (k Keeper) SetDelegationQueueSlice(ctx *sdk.Context, tasks []types.DelegationTask, height uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DelegationTasks{DelegationTasks: tasks})
	store.Set(types.GetDelegateQueueKey(height), bz)
}

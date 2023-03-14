package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"

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
	ibcTransferKeeper   ibctransferkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:            storeKey,
		cdc:                 cdc,
		authority:           authority,
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		icaControllerKeeper: icaControllerKeeper,
		ibcTransferKeeper:   ibcTransferKeeper,
	}
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

	sourceChainMetadata = &types.SourceChainMetadata{}

	err := k.cdc.Unmarshal(bz, sourceChainMetadata)
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

func (k Keeper) PushDelegationTaskQueue(ctx *sdk.Context, queueKey []byte, delegationTask *types.DelegationTask) {
	tasks := k.GetDelegationQueueSlice(ctx, queueKey, uint64(ctx.BlockHeight()))

	height := uint64(ctx.BlockHeight())
	if len(tasks) == 0 {
		k.SetDelegationQueueSlice(ctx, queueKey, []types.DelegationTask{*delegationTask}, height)
	} else {
		tasks = append(tasks, *delegationTask)
		k.SetDelegationQueueSlice(ctx, queueKey, tasks, height)
	}
}

func (k Keeper) GetDelegationQueueSlice(ctx *sdk.Context, queueKey []byte, height uint64) []types.DelegationTask {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDelegateQueueKey(queueKey, height))

	if bz == nil {
		return []types.DelegationTask{}
	}

	tasks := types.DelegationTasks{}
	k.cdc.MustUnmarshal(bz, &tasks)

	return tasks.DelegationTasks
}

func (k Keeper) SetDelegationQueueSlice(ctx *sdk.Context, queueKey []byte, tasks []types.DelegationTask, height uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DelegationTasks{DelegationTasks: tasks})
	store.Set(types.GetDelegateQueueKey(queueKey, height), bz)
}

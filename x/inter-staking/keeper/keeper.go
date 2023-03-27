package keeper

import (
	"time"

	"celinium/x/inter-staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibcclientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
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
	ibcClientKeeper     ibcclientkeeper.Keeper
	scopedKeeper        capabilitykeeper.ScopedKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	ibcClientKeeper ibcclientkeeper.Keeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	return Keeper{
		storeKey:            storeKey,
		cdc:                 cdc,
		authority:           authority,
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		icaControllerKeeper: icaControllerKeeper,
		ibcTransferKeeper:   ibcTransferKeeper,
		ibcClientKeeper:     ibcClientKeeper,
		scopedKeeper:        scopedKeeper,
	}
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
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

func (k Keeper) PushDelegationTaskQueue(ctx *sdk.Context, queueKey []byte, sequence uint64, delegationTask *types.DelegationTask) {
	tasks := k.GetDelegationQueueSlice(ctx, queueKey, sequence)

	if len(tasks) == 0 {
		k.SetDelegationQueueSlice(ctx, queueKey, []types.DelegationTask{*delegationTask}, sequence)
	} else {
		tasks = append(tasks, *delegationTask)
		k.SetDelegationQueueSlice(ctx, queueKey, tasks, sequence)
	}
}

func (k Keeper) GetDelegationQueueSlice(ctx *sdk.Context, queueKey []byte, sequence uint64) []types.DelegationTask {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDelegateQueueKey(queueKey, sequence))

	if bz == nil {
		return []types.DelegationTask{}
	}

	tasks := types.DelegationTasks{}
	k.cdc.MustUnmarshal(bz, &tasks)

	return tasks.DelegationTasks
}

func (k Keeper) SetDelegationQueueSlice(ctx *sdk.Context, queueKey []byte, tasks []types.DelegationTask, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DelegationTasks{DelegationTasks: tasks})
	store.Set(types.GetDelegateQueueKey(queueKey, sequence), bz)
}

func (k Keeper) SetDelegation(ctx *sdk.Context, delegator string, chainID string, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&amount)

	delegationKey := types.GetDelegationKey(delegator, chainID)
	if amountBz := store.Get(delegationKey); amountBz != nil {
		var coin sdk.Coin

		k.cdc.MustUnmarshal(amountBz, &coin)
		coin.Amount.Add(amount.Amount)
		bz := k.cdc.MustMarshal(&coin)

		store.Set(delegationKey, bz)
		return
	}

	store.Set(delegationKey, bz)
}

func (k Keeper) GetDelegation(ctx sdk.Context, delegator string, chainID string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetDelegationKey(delegator, chainID))
	var coin sdk.Coin
	k.cdc.MustUnmarshal(bz, &coin)

	return coin
}

func (k Keeper) SetPreparingUndelegation(ctx sdk.Context, completeTime time.Time, sequence uint64, chainID string, delegator string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPreparingUndelegationKey(completeTime, sequence)

	// use anthother type ??
	undelegation := types.DelegationTask{
		ChainId:   chainID,
		Delegator: delegator,
		Amount:    sdk.Coin{},
	}

	bz := k.cdc.MustMarshal(&undelegation)

	store.Set(key, bz)
}

func (k Keeper) GetSourceChainUnbondingDelegations(ctx sdk.Context, chainID, clientID string) *types.SourceChainUnbondingQueue {
	store := ctx.KVStore(k.storeKey)

	key := types.GetSourceChainUnbondingDelegationsKey(chainID, clientID)
	if !store.Has(key) {
		return nil
	}

	bz := store.Get(key)

	var ubd types.SourceChainUnbondingQueue

	err := k.cdc.Unmarshal(bz, &ubd)
	if err != nil {
		panic(err)
	}

	return &ubd
}

func (k Keeper) SetSourceChainUnbondingDelegations(ctx sdk.Context, ubd types.SourceChainUnbondingQueue) {
	bz := k.cdc.MustMarshal(&ubd)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetSourceChainUnbondingDelegationsKey(ubd.ChainID, ubd.ClientID), bz)
}

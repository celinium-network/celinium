package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (k Keeper) AddSouceChain(ctx sdk.Context, sourceChain *types.SourceChain) error {
	if err := sourceChain.BasicVerify(); err != nil {
		return sdkerrors.Wrapf(types.ErrSourceChainParameter, "error: %v", err)
	}

	// check source chain wheather is already existed.
	if _, found := k.GetSourceChain(ctx, sourceChain.ChainID); found {
		return sdkerrors.Wrapf(types.ErrSourceChainExist, "already exist source chain, ID: %s", sourceChain.ChainID)
	}

	accounts := sourceChain.GenerateAndFillAccount(ctx)

	for _, a := range accounts {
		k.accountKeeper.NewAccount(ctx, a)
		k.accountKeeper.SetAccount(ctx, a)
		if err := k.icaCtlKeeper.RegisterInterchainAccount(ctx, sourceChain.ConnectionID, a.GetAddress().String(), ""); err != nil {
			return err
		}
	}

	k.SetSourceChain(ctx, sourceChain)

	return nil
}

// CreateDelegationRecordForEpoch create a new DelegationRecord in current epoch for all available chain.
// If current epoch already has DelegationRecord for the chain, then do nothing
func (k Keeper) CreateDelegationRecordForEpoch(ctx sdk.Context, epochNumber int64) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	sourcechain := &types.SourceChain{}
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		if _, found := k.GetChianDelegationRecordID(ctx, sourcechain.ChainID, uint64(epochNumber)); found {
			continue
		}

		id := k.GetDelegationRecordID(ctx)

		record := types.DelegationRecord{}

		k.SetChainDelegationRecordID(ctx, sourcechain.ChainID, uint64(epochNumber), id)

		k.SetDelegationRecord(ctx, id, &record)
	}
}

// CreateEpochUnbondings a new unbonding in current epoch.
func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber int64) {
	_, found := k.GetEpochUnboundings(ctx, uint64(epochNumber))
	if found {
		return
	}

	epochUnbonding := types.EpochUnbondings{
		Epoch:      uint64(epochNumber),
		Unbondings: []types.Unbonding{},
	}

	k.SetEpochUnboundings(ctx, &epochUnbonding)
}

// GetDelegationRecord return DelegationRecord by id
func (k Keeper) GetDelegationRecord(ctx sdk.Context, id uint64) (*types.DelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetDelegationRecordKey(id))
	if bz == nil {
		return nil, false
	}

	record := &types.DelegationRecord{}
	k.cdc.MustUnmarshal(bz, record)

	return record, true
}

// SetDelegationRecord store DelegationRecord
func (k Keeper) SetDelegationRecord(ctx sdk.Context, id uint64, record *types.DelegationRecord) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(record)

	store.Set(types.GetDelegationRecordKey(id), bz)
}

// GetChianDelegationRecordID get DelegationRecord's ID of a chain by epoch and chainID
func (k Keeper) GetChianDelegationRecordID(ctx sdk.Context, chainID string, epochNumber uint64) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GeChainDelegationRecordIDForEpochKey(epochNumber, []byte(chainID)))

	if len(bz) == 0 {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

// SetChianDelegationRecordID set DelegationRecord's ID  of chain at specific epoch and chainID
func (k Keeper) SetChainDelegationRecordID(ctx sdk.Context, chainID string, epochNumber uint64, recordID uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GeChainDelegationRecordIDForEpochKey(epochNumber, []byte(chainID)), sdk.Uint64ToBigEndian(recordID))
}

// chainAvaiable wheather a chain is available. when all interchain account is registered, then it's available
func (k Keeper) sourceChainAvaiable(ctx sdk.Context, sourceChain *types.SourceChain) bool {
	_, found1 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
	_, found2 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	_, found3 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.UnboudAddress)

	if found1 && found2 && found3 {
		return true
	}

	return false
}

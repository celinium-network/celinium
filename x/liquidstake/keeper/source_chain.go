package keeper

import (
	"strings"

	sdkerrors "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"

	transfertype "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

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

	parts := []string{transfertype.PortID, sourceChain.TransferChannelID, sourceChain.NativeDenom}
	denom := strings.Join(parts, "/")
	denomTrace := transfertype.ParseDenomTrace(denom)
	sourceChain.IbcDenom = denomTrace.IBCDenom()

	icaAccounts := sourceChain.GenerateAccounts(ctx)

	for _, a := range icaAccounts {
		k.accountKeeper.NewAccount(ctx, a)
		k.accountKeeper.SetAccount(ctx, a)
		if err := k.icaCtlKeeper.RegisterInterchainAccount(ctx, sourceChain.ConnectionID, a.GetAddress().String(), ""); err != nil {
			return err
		}
	}

	k.SetSourceChain(ctx, sourceChain)

	return nil
}

// CreateEpochDelegationRecord create a new DelegationRecord in current epoch for all available chain.
// If current epoch already has DelegationRecord for the chain, then do nothing
func (k Keeper) CreateEpochDelegationRecord(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	sourcechain := &types.SourceChain{}
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		if _, found := k.GetChianDelegationRecordID(ctx, sourcechain.ChainID, epochNumber); found {
			continue
		}

		id := k.GetDelegationRecordID(ctx)

		record := types.DelegationRecord{
			Id:             id,
			DelegationCoin: sdk.NewCoin(sourcechain.IbcDenom, sdk.ZeroInt()),
			Status:         types.DelegationPending,
			EpochNumber:    epochNumber,
			ChainID:        sourcechain.ChainID,
		}

		k.SetChainDelegationRecordID(ctx, sourcechain.ChainID, epochNumber, id)

		k.SetDelegationRecord(ctx, id, &record)

		k.IncreaseDelegationRecordID(ctx)
	}
}

// CreateEpochUnbondings a new unbonding in current epoch.
func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber uint64) {
	_, found := k.GetEpochUnboundings(ctx, epochNumber)
	if found {
		return
	}

	epochUnbonding := types.EpochUnbondings{
		Epoch:      epochNumber,
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
	findICA := func(addr string) bool {
		portID, err := icatypes.NewControllerPortID(addr)
		if err != nil {
			return false
		}
		_, found := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, portID)
		return found
	}

	if findICA(sourceChain.WithdrawAddress) && findICA(sourceChain.DelegateAddress) {
		return true
	}

	return false
}

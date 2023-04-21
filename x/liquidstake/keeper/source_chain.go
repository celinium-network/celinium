package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"

	appparams "github.com/celinium-netwok/celinium/app/params"
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

	if err := sourceChain.GenerateIBCDeonm(); err != nil {
		return sdkerrors.Wrapf(types.ErrSourceChainParameter, err.Error())
	}

	connection, found := k.ibcKeeper.ConnectionKeeper.GetConnection(ctx, sourceChain.ConnectionID)
	if !found {
		return sdkerrors.Wrapf(types.ErrSourceChainParameter, "connection not find: ID %s", sourceChain.ConnectionID)
	}

	icaAccounts := sourceChain.GenerateAccounts(ctx)
	icaVersion := icatypes.ModuleCdc.MustMarshalJSON((&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: sourceChain.ConnectionID,
		HostConnectionId:       connection.Counterparty.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	for _, a := range icaAccounts {
		k.accountKeeper.NewAccount(ctx, a)
		k.accountKeeper.SetAccount(ctx, a)
		if err := k.icaCtlKeeper.RegisterInterchainAccount(ctx, sourceChain.ConnectionID, a.GetAddress().String(), string(icaVersion)); err != nil {
			return err
		}
	}

	delegatieEpochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", appparams.DelegationEpochIdentifier)
	}

	k.createChainEpochDelegationRecord(ctx, uint64(delegatieEpochInfo.CurrentEpoch), sourceChain.ChainID, sourceChain.IbcDenom)

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

		k.createChainEpochDelegationRecord(ctx, epochNumber, sourcechain.ChainID, sourcechain.IbcDenom)
	}
}

func (k Keeper) createChainEpochDelegationRecord(ctx sdk.Context, epochNumber uint64, chainID string, stakeDenom string) *types.DelegationRecord {
	id := k.GetDelegationRecordID(ctx)

	record := types.DelegationRecord{
		Id:             id,
		DelegationCoin: sdk.NewCoin(stakeDenom, sdk.ZeroInt()),
		Status:         types.DelegationPending,
		EpochNumber:    epochNumber,
		ChainID:        chainID,
	}

	k.SetChainDelegationRecordID(ctx, chainID, epochNumber, id)

	k.SetDelegationRecord(ctx, id, &record)

	k.IncreaseDelegationRecordID(ctx)

	return &record
}

// CreateEpochUnbondings a new unbonding in current epoch.
func (k Keeper) CreateEpochUnbondings(ctx sdk.Context, epochNumber uint64) *types.EpochUnbondings {
	_, found := k.GetEpochUnboundings(ctx, epochNumber)
	if found {
		return nil
	}

	epochUnbonding := types.EpochUnbondings{
		Epoch:      epochNumber,
		Unbondings: []types.Unbonding{},
	}

	k.SetEpochUnboundings(ctx, &epochUnbonding)

	return &epochUnbonding
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

func (k Keeper) GetSourceChainAddr(ctx sdk.Context, connectionID string, ctlAddress string) (string, error) {
	portID, err := icatypes.NewControllerPortID(ctlAddress)
	if err != nil {
		return "", err
	}

	sourceChainAddr, found := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, connectionID, portID)
	if !found {
		return "", sdkerrors.Wrapf(types.ErrICANotFound, "connectionID %s ctlAddress %s", connectionID, ctlAddress)
	}
	return sourceChainAddr, nil
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

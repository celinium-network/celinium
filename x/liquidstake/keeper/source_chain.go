package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
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
		if err := k.icaCtlKeeper.RegisterInterchainAccount(ctx, sourceChain.ConnectionID,
			a.GetAddress().String(), string(icaVersion)); err != nil {
			return err
		}
	}

	delegatieEpochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s",
			appparams.DelegationEpochIdentifier)
	}

	if _, err := k.createProxyDelegation(ctx, uint64(delegatieEpochInfo.CurrentEpoch),
		sourceChain.ChainID, sourceChain.IbcDenom); err != nil {
		return err
	}

	k.SetSourceChain(ctx, sourceChain)

	return nil
}

// CreateProxyDelegationForEpoch create a new ProxyDelegation in current epoch for all available chain.
// If current epoch already has ProxyDelegation for the chain, then do nothing
func (k Keeper) CreateProxyDelegationForEpoch(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	sourcechain := &types.SourceChain{}
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		if _, found := k.GetChianProxyDelegationID(ctx, sourcechain.ChainID, epochNumber); found {
			continue
		}

		k.createProxyDelegation(ctx, epochNumber, sourcechain.ChainID, sourcechain.IbcDenom)
	}
}

func (k Keeper) createProxyDelegation(ctx sdk.Context, epochNumber uint64, chainID string, stakeDenom string) (*types.ProxyDelegation, error) {
	id := k.GetProxyDelegationID(ctx)

	delegation := types.ProxyDelegation{
		Id:          id,
		Coin:        sdk.NewCoin(stakeDenom, sdk.ZeroInt()),
		Status:      types.ProxyDelegationPending,
		EpochNumber: epochNumber,
		ChainID:     chainID,
	}

	if err := k.IncreaseProxyDelegationID(ctx); err != nil {
		// uint64 exhausted, shoudle change id store prefix of `ProxyDelegation`
		return nil, err
	}

	k.SetChainProxyDelegationID(ctx, chainID, epochNumber, id)

	k.SetProxyDelegation(ctx, id, &delegation)

	return &delegation, nil
}

// CreateProxyUnbondingForEpoch a new unbonding in current epoch.
func (k Keeper) CreateProxyUnbondingForEpoch(ctx sdk.Context, epochNumber uint64) *types.EpochProxyUnbonding {
	_, found := k.GetEpochProxyUnboundings(ctx, epochNumber)
	if found {
		return nil
	}

	epochUnbonding := types.EpochProxyUnbonding{
		Epoch:      epochNumber,
		Unbondings: []types.ProxyUnbonding{},
	}

	k.SetEpochProxyUnboundings(ctx, &epochUnbonding)

	return &epochUnbonding
}

// GetProxyDelegation return ProxyDelegation by id
func (k Keeper) GetProxyDelegation(ctx sdk.Context, id uint64) (*types.ProxyDelegation, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetProxyDelegationKey(id))
	if bz == nil {
		return nil, false
	}

	delegation := &types.ProxyDelegation{}
	k.cdc.MustUnmarshal(bz, delegation)

	return delegation, true
}

// SetProxyDelegation store ProxyDelegation
func (k Keeper) SetProxyDelegation(ctx sdk.Context, id uint64, delegation *types.ProxyDelegation) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(delegation)

	store.Set(types.GetProxyDelegationKey(id), bz)
}

// GetChianProxyDelegationID get ProxyDelegation's ID of a chain by epoch and chainID
func (k Keeper) GetChianProxyDelegationID(ctx sdk.Context, chainID string, epochNumber uint64) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetChainProxyDelegationIDForEpochKey(epochNumber, []byte(chainID)))

	if len(bz) == 0 {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

// SetChianProxyDelegationID set ProxyDelegation's ID  of chain at specific epoch and chainID
func (k Keeper) SetChainProxyDelegationID(ctx sdk.Context, chainID string, epochNumber uint64, delegationID uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetChainProxyDelegationIDForEpochKey(epochNumber, []byte(chainID)), sdk.Uint64ToBigEndian(delegationID))
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

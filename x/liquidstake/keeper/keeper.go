package keeper

import (
	sdkioerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	epochKeeper       types.EpochKeeper
	ibcKeeper         *ibckeeper.Keeper
	ibcTransferKeeper ibctransferkeeper.Keeper
	icaCtlKeeper      icacontrollerkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	epochKeeper types.EpochKeeper,
	ibcClientKeeper *ibckeeper.Keeper,
	icaCtlKeeper icacontrollerkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		epochKeeper:       epochKeeper,
		ibcKeeper:         ibcClientKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		icaCtlKeeper:      icaCtlKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetSourceChain get source chain by chainID
func (k Keeper) GetSourceChain(ctx sdk.Context, chainID string) (*types.SourceChain, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetSourceChainKey([]byte(chainID)))
	if bz == nil {
		return nil, false
	}

	sourceChain := &types.SourceChain{}
	if err := k.cdc.Unmarshal(bz, sourceChain); err != nil {
		return nil, false
	}

	return sourceChain, true
}

func (k Keeper) SetSourceChain(ctx sdk.Context, sourceChain *types.SourceChain) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(sourceChain)

	store.Set(types.GetSourceChainKey([]byte(sourceChain.ChainID)), bz)
}

func (k Keeper) GetProxyDelegationID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ProxyDelegationIDKey)

	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) IncreaseProxyDelegationID(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProxyDelegationIDKey)

	oldID := sdk.BigEndianToUint64(bz)
	newID := oldID + 1
	if newID < oldID {
		return sdkioerrors.Wrapf(sdk.ErrIntOverflowAbci, "failed to increase proxy delegation ID")
	}

	store.Set(types.ProxyDelegationIDKey, sdk.Uint64ToBigEndian(newID))

	return nil
}

func (k Keeper) GetAllProxyDelegation(ctx sdk.Context) []types.ProxyDelegation {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ProxyDelegationPrefix)

	var records []types.ProxyDelegation
	for ; iterator.Valid(); iterator.Next() {
		r := types.ProxyDelegation{}

		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, &r)
		records = append(records, r)
	}

	return records
}

// checkIBCClient check weather the ibcclient of the specific chain is active
// func (k Keeper) checkIBCClient(ctx sdk.Context, chainID string) error {
// 	clientState, found := k.ibcClientKeeper.GetClientState(ctx, chainID)
// 	if !found {
// 		return sdkerrors.Wrapf(ibcclienttypes.ErrClientNotFound, "unknown client, ID: %s", chainID)
// 	}

// 	clientStore := k.ibcClientKeeper.ClientStore(ctx, chainID)

// 	if status := clientState.Status(ctx, clientStore, k.cdc); status != exported.Active {
// 		return sdkerrors.Wrapf(ibcclienttypes.ErrClientNotActive, "cannot update client (%s) with status %s", chainID, status)
// 	}

// 	return nil
// }

// sendCoinsFromAccountToAccount preform send coins form sender to receiver.
func (k Keeper) sendCoinsFromAccountToAccount(ctx sdk.Context, senderAddr sdk.AccAddress, receiverAddr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, amt); err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddr, amt)
}

func (k Keeper) mintCoins(ctx sdk.Context, receiverAddr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amt); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddr, amt)
}

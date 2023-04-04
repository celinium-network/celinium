package keeper

import (
	"celinium/x/liquidstake/types"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibcclientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper     types.AccountKeeper
	bankKeeper        types.BankKeeper
	epochKeeper       types.EpochKeeper
	ibcClientKeeper   ibcclientkeeper.Keeper
	ibcTransferKeeper ibctransferkeeper.Keeper
	icaCtlKeeper      icacontrollerkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	epochKeeper types.EpochKeeper,
	ibcClientKeeper ibcclientkeeper.Keeper,
	icaCtlKeeper icacontrollerkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		epochKeeper:       epochKeeper,
		ibcClientKeeper:   ibcClientKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		icaCtlKeeper:      icaCtlKeeper,
	}
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

func (k Keeper) GetDelegationRecordID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.DelegationRecordIDKey)

	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) IncreaseDelegationRecordID(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DelegationRecordIDKey)

	oldID := sdk.BigEndianToUint64(bz)
	oldID++ // todo! need check overflow?

	store.Set(types.DelegationRecordIDKey, sdk.Uint64ToBigEndian(oldID))
}

// checkIBCClient check weather the ibcclient of the specific chain is active
func (k Keeper) checkIBCClient(ctx sdk.Context, chainID string) error {
	clientState, found := k.ibcClientKeeper.GetClientState(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(ibcclienttypes.ErrClientNotFound, "unknown client, ID: %s", chainID)
	}

	clientStore := k.ibcClientKeeper.ClientStore(ctx, chainID)

	if status := clientState.Status(ctx, clientStore, k.cdc); status != exported.Active {
		return sdkerrors.Wrapf(ibcclienttypes.ErrClientNotActive, "cannot update client (%s) with status %s", chainID, status)
	}

	return nil
}

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

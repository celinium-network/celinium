package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibcclientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"celinium/x/liquid-stake/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper types.AccountKeeper

	ibcClientKeeper   ibcclientkeeper.Keeper
	ibcTransferKeeper ibctransferkeeper.Keeper
	icaCtlKeeper      icacontrollerkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	ibcClientKeeper ibcclientkeeper.Keeper,
	icaCtlKeeper icacontrollerkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		accountKeeper:     accountKeeper,
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

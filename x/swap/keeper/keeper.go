package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/tendermint/tendermint/libs/log"

	"celinium/x/swap/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	paramSpace paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramSpace: ps,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) CreateModuleAccount(ctx sdk.Context) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter, authtypes.Burner)
	k.accountKeeper.SetModuleAccount(ctx, moduleAcc)
}

func (k Keeper) GetNextPairId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextPairId := gogotypes.UInt64Value{}

	b := store.Get(types.KeyNextGlobalPairId)
	if b == nil {
		panic(fmt.Errorf("getting at key (%v) should not have been nil", types.KeyNextGlobalPairId))
	}
	if err := proto.Unmarshal(b, &nextPairId); err != nil {
		panic(err)
	}

	return nextPairId.Value
}

func (k Keeper) SetNextPairId(ctx sdk.Context, poolId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&gogotypes.UInt64Value{Value: poolId})
	if err != nil {
		panic(err)
	}

	store.Set(types.KeyNextGlobalPairId, bz)
}

func (k Keeper) SetTokensToPoolId(ctx sdk.Context, token0 string, token1 string, poolId uint64) {
	prefix := types.GetKeyPrefixTokensToPoolId(token0, token1)

	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&gogotypes.UInt64Value{Value: poolId})
	if err != nil {
		panic(err)
	}

	store.Set(prefix, bz)
}

func (k Keeper) GetPoolIdFromTokens(ctx sdk.Context, token0 string, token1 string) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetKeyPrefixTokensToPoolId(token0, token1)

	bz := store.Get(prefix)
	pairId := gogotypes.UInt64Value{}

	if len(bz) == 0 {
		return 0, types.ErrPairNotExist
	}

	if err := proto.Unmarshal(bz, &pairId); err != nil {
		return 0, err
	}
	return pairId.Value, nil
}

func (k Keeper) SetIdToPair(ctx sdk.Context, pairId uint64, pair *types.Pair) {
	bz, err := k.cdc.Marshal(pair)
	if err != nil {
		panic(err)
	}
	store := ctx.KVStore(k.storeKey)

	pairKey := types.GetKeyPrefixPairs(pairId)
	store.Set(pairKey, bz)
}

func (k Keeper) GetPairFromId(ctx sdk.Context, pairId uint64) *types.Pair {
	store := ctx.KVStore(k.storeKey)

	pairKey := types.GetKeyPrefixPairs(pairId)
	bz := store.Get(pairKey)

	pair := types.Pair{}
	if err := proto.Unmarshal(bz, &pair); err != nil {
		panic(err)
	}
	return &pair
}

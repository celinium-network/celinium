package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	epochKeeper   types.EpochKeeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetMultiStakingDenomWhiteList(ctx sdk.Context) (*types.MultiStakingDenomWhiteList, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.MultiStakingDenomWhiteListKey)
	if bz == nil {
		return nil, false
	}

	whiteList := &types.MultiStakingDenomWhiteList{}

	if err := k.cdc.Unmarshal(bz, whiteList); err != nil {
		return nil, false
	}

	return whiteList, true
}

func (k Keeper) SetMultiStakingDenom(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMultiStakingDenomWhiteList(ctx)
	if !found || whiteList == nil {
		whiteList = &types.MultiStakingDenomWhiteList{
			DenomList: []string{denom},
		}
	} else {
		for _, existedDenom := range whiteList.DenomList {
			if strings.Compare(existedDenom, denom) == 0 {
				return false
			}
		}

		whiteList.DenomList = append(whiteList.DenomList, denom)
	}

	bz, err := k.cdc.Marshal(whiteList)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.MultiStakingDenomWhiteListKey, bz)

	return true
}

func (k Keeper) GetMultiStakingAgentByID(ctx sdk.Context, id uint64) (*types.MultiStakingAgent, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMultiStakingAgentKey(id))

	if bz == nil {
		return nil, false
	}

	agent := &types.MultiStakingAgent{}
	k.cdc.MustUnmarshal(bz, agent)

	return agent, true
}

func (k Keeper) GetMultiStakingAgent(ctx sdk.Context, denom string, valAddr string) (*types.MultiStakingAgent, bool) {
	agentID, found := k.GetMultiStakingAgentIDByDenomAndVal(ctx, denom, valAddr)
	if !found {
		return nil, false
	}

	return k.GetMultiStakingAgentByID(ctx, agentID)
}

func (k Keeper) SetMultiStakingAgent(ctx sdk.Context, agent *types.MultiStakingAgent) uint64 {
	latestAgentID := k.GetLatestMultiStakingAgentID(ctx)
	latestAgentID++

	bz := k.cdc.MustMarshal(agent)
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetMultiStakingAgentKey(latestAgentID), bz)

	k.SetMultiStakingAgentIDByDenomAndVal(ctx, latestAgentID, agent.StakeDeonm, agent.AgentDelegatorAddress)
	k.SetLatestMultiStakingAgentID(ctx, latestAgentID)

	return latestAgentID
}

func (k Keeper) GetMultiStakingAgentIDByDenomAndVal(ctx sdk.Context, denom string, valAddr string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMultiStakingAgentIDKey(denom, valAddr))
	if bz == nil {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetMultiStakingAgentIDByDenomAndVal(ctx sdk.Context, id uint64, denom, valAddr string) {
	store := ctx.KVStore(k.storeKey)
	idBz := sdk.Uint64ToBigEndian(id)

	store.Set(types.GetMultiStakingAgentIDKey(denom, valAddr), idBz)
}

func (k Keeper) GetLatestMultiStakingAgentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.MultiStakingLatestAgentIDKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLatestMultiStakingAgentID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	idBz := sdk.Uint64ToBigEndian(id)

	store.Set(types.MultiStakingLatestAgentIDKey, idBz)
}

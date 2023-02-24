package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"celinium/x/swap/types"
)

func (k Keeper) createPair(ctx sdk.Context, token0, token1 string) (*types.Pair, error) {
	if _, exist := k.pairExisted(ctx, token0, token1); exist {
		return nil, types.ErrPairCreated
	}

	if !k.bankKeeper.HasSupply(ctx, token0) {
		return nil, fmt.Errorf("%s has't supply", token0)
	}

	if !k.bankKeeper.HasSupply(ctx, token1) {
		return nil, fmt.Errorf("%s has't supply", token1)
	}

	pairId := k.getNextPairIdAndIncrement(ctx)

	pair, err := types.CreatePair(pairId, token0, token1)
	if err != nil {
		return nil, err
	}

	k.bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
		Description: fmt.Sprintf("the lp token of the swap pair%d", pairId),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    pair.LpToken.Denom,
				Exponent: 0,
			},
		},
		Base:    pair.LpToken.Denom,
		Display: pair.LpToken.Denom,
	})

	k.SetIdToPair(ctx, pairId, pair)

	k.SetTokensToPoolId(ctx, pair.Token0.Denom, pair.Token1.Denom, pairId)

	return pair, nil
}

func (k Keeper) getNextPairIdAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPairId(ctx)
	k.SetNextPairId(ctx, nextPoolId+1)
	return nextPoolId
}

func (k Keeper) pairExisted(ctx sdk.Context, token0, token1 string) (uint64, bool) {
	sortToken0, sortToken1 := types.SortToken(token0, token1)
	id, err := k.GetPoolIdFromTokens(ctx, sortToken0, sortToken1)
	return id, err != types.ErrPairNotExist
}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, coin sdk.Coin) (sdk.Coin, error) {
	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)

	return k.EquivalentCoinCalculator(ctx, coin, defaultBondDenom)
}

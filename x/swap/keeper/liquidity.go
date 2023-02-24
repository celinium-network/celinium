package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/swap/types"
)

type addLiqudityResult struct {
	Token0  *sdk.Coin
	Token1  *sdk.Coin
	TpToken *sdk.Coin
}

func (k Keeper) addLiquidty(
	ctx sdk.Context,
	sender sdk.AccAddress,
	token0 *sdk.Coin,
	token1 *sdk.Coin,
	amount0Min sdkmath.Int,
	amount1Min sdkmath.Int,
	receipt sdk.AccAddress,
) (*addLiqudityResult, error) {
	pairId, exist := k.pairExisted(ctx, token0.Denom, token1.Denom)
	if !exist {
		return nil, types.ErrPairNotExist
	}

	pair := k.GetPairFromId(ctx, pairId)
	pairAccount := sdk.MustAccAddressFromBech32(pair.Account)

	reserve0 := k.bankKeeper.GetBalance(ctx, pairAccount, token0.Denom)
	reserve1 := k.bankKeeper.GetBalance(ctx, pairAccount, token1.Denom)

	amount0, amount1, err := calculate_added_amount(&token0.Amount, &token1.Amount, &amount0Min, &amount1Min, &reserve0.Amount, &reserve1.Amount)
	if err != nil {
		return nil, err
	}

	balance0 := k.bankKeeper.GetBalance(ctx, sender, token0.Denom)
	balance1 := k.bankKeeper.GetBalance(ctx, sender, token1.Denom)
	totalLiqudity := k.bankKeeper.GetSupply(ctx, pair.LpToken.Denom)

	if balance0.Amount.LT(*amount0) || balance1.Amount.LT(*amount1) {
		return nil, types.ErrInsufficientFunds
	}

	// todo! collect fee

	mintedLiquidity := calculate_liquidity(amount0, amount1, &reserve0.Amount, &reserve1.Amount, &totalLiqudity.Amount)
	if mintedLiquidity.IsZero() {
		return nil, types.ErrMath
	}

	pair.Token0 = pair.Token0.AddAmount(*amount0)
	pair.Token1 = pair.Token1.AddAmount(*amount1)
	pair.LpToken = pair.LpToken.AddAmount(*mintedLiquidity)

	mintedLiquidityToken := sdk.Coins{sdk.NewCoin(pair.LpToken.Denom, *mintedLiquidity)}
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintedLiquidityToken)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receipt, mintedLiquidityToken)
	if err != nil {
		return nil, err
	}

	transferTokens := sdk.Coins{sdk.NewCoin(pair.Token0.Denom, *amount0), sdk.NewCoin(pair.Token1.Denom, *amount1)}
	err = k.bankKeeper.SendCoins(ctx, sender, pairAccount, transferTokens)
	if err != nil {
		return nil, err
	}

	// update pair
	k.SetIdToPair(ctx, pairId, pair)

	return &addLiqudityResult{
		Token0:  &transferTokens[0],
		Token1:  &transferTokens[1],
		TpToken: &mintedLiquidityToken[0],
	}, nil
}

func calculate_added_amount(
	amount0Desired *sdkmath.Int,
	amount1Desired *sdkmath.Int,
	amount0Min *sdkmath.Int,
	amount1Min *sdkmath.Int,
	reserve0 *sdkmath.Int,
	reserve1 *sdkmath.Int,
) (*sdkmath.Int, *sdkmath.Int, error) {
	if reserve0.IsZero() || reserve1.IsZero() {
		return amount0Desired, amount1Desired, nil
	}
	amount1Optimal := MulDiv(amount0Desired, reserve0, reserve1)
	if amount1Optimal.LTE(*amount1Desired) {
		if amount1Optimal.LT(*amount1Min) {
			return nil, nil, types.ErrInvalidTokenAmountRange
		}
		return amount0Desired, amount1Optimal, nil
	}

	amount0Optimal := MulDiv(amount1Desired, reserve1, reserve0)
	if amount0Optimal.LTE(*amount0Desired) {
		if amount0Optimal.LT(*amount0Min) {
			return nil, nil, types.ErrInvalidTokenAmountRange
		}
		return amount0Optimal, amount1Desired, nil
	}

	return amount0Desired, amount1Desired, nil
}

func calculate_liquidity(
	amount0 *sdkmath.Int,
	amount1 *sdkmath.Int,
	reserve0 *sdkmath.Int,
	reserve1 *sdkmath.Int,
	total_liquidity *sdkmath.Int,
) *sdkmath.Int {
	if total_liquidity.IsZero() {
		res := sdkmath.NewIntFromBigInt(big.NewInt(0).Sqrt(amount0.Mul(*amount1).BigInt()))
		return &res
	}
	c0 := MulDiv(amount0, reserve0, total_liquidity)
	c1 := MulDiv(amount1, reserve1, total_liquidity)

	if c0.GT(*c1) {
		return c1
	}

	return c0
}

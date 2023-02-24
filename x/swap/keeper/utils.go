package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

// MulDiv a * b /c
func MulDiv(a *sdkmath.Int, b *sdkmath.Int, c *sdkmath.Int) *sdkmath.Int {
	res := sdkmath.NewIntFromBigInt(big.NewInt(0).Div(a.Mul(*b).BigInt(), c.BigInt()))
	return &res
}

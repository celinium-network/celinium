package types

import (
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleLpPrefix = "swapLp"
)

func CreatePair(pairId uint64, token0, token1 string) (*Pair, error) {
	if token0 == token1 {
		return nil, ErrInvalidTokens
	}

	sortToken0, sortToken1 := SortToken(token0, token1)
	lp_token, err := getPairLpToken(sortToken0, sortToken1)
	if err != nil {
		return nil, err
	}

	return &Pair{
		Account: NewPoolAddress(pairId).String(),
		Token0: sdk.Coin{
			Denom:  sortToken0,
			Amount: math.ZeroInt(),
		},
		Token1: sdk.Coin{
			Denom:  sortToken1,
			Amount: math.ZeroInt(),
		},
		LpToken: sdk.Coin{
			Denom:  lp_token,
			Amount: math.ZeroInt(),
		},
	}, nil
}

// getPairLpToken constructs a lp token
// The pair lp token constructed is swap/{token0}/{token1}
func getPairLpToken(token0, token1 string) (string, error) {
	lptoken := strings.Join([]string{ModuleLpPrefix, token0, token1}, "/")
	return lptoken, sdk.ValidateDenom(lptoken)
}

func SortToken(token0, token1 string) (string, string) {
	if token0 > token1 {
		return token1, token0
	}
	return token0, token1
}

func NewPoolAddress(poolId uint64) sdk.AccAddress {
	key := append([]byte("pair"), sdk.Uint64ToBigEndian(poolId)...)
	return address.Module(ModuleName, key)
}

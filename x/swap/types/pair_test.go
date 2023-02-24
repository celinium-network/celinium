package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	appparams "celinium/app/params"
	"celinium/x/swap/types"
)

func TestCreatePair(t *testing.T) {
	appparams.SetAddressPrefixes()
	appparams.SetAddressPrefixes()

	for _, tc := range []struct {
		desc            string
		pairId          uint64
		token0          string
		token1          string
		expectedToken0  string
		expectedToken1  string
		expectedLpToken string
		expectedAccount string
		err             error
	}{
		{
			desc:            "same tokens invalid",
			pairId:          0,
			token0:          "token0",
			token1:          "token0",
			expectedToken0:  "",
			expectedToken1:  "",
			expectedLpToken: "expectedLpToken",
			expectedAccount: "",
			err:             types.ErrInvalidTokens,
		},
		{
			desc:            "normal",
			pairId:          0,
			token0:          "token0",
			token1:          "token1",
			expectedToken0:  "token0",
			expectedToken1:  "token1",
			expectedLpToken: "swapLp/token0/token1",
			expectedAccount: "",
			err:             nil,
		},
		{
			desc:            "should sorted tokens",
			pairId:          0,
			token0:          "token1",
			token1:          "token0",
			expectedToken0:  "token0",
			expectedToken1:  "token1",
			expectedLpToken: "swapLp/token0/token1",
			expectedAccount: "",
			err:             nil,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			createdPair, err := types.CreatePair(tc.pairId, tc.token0, tc.token1)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				require.Nil(t, createdPair)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedToken0, createdPair.Token0.Denom)
				require.Equal(t, sdkmath.ZeroInt(), createdPair.Token0.Amount)
				require.Equal(t, tc.expectedToken1, createdPair.Token1.Denom)
				require.Equal(t, sdkmath.ZeroInt(), createdPair.Token1.Amount)
				require.Equal(t, tc.expectedLpToken, createdPair.LpToken.Denom)
			}
		})
	}
}

package keeper_test

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/swap/types"
)

func (suite *IntegrationTestSuite) createTestToken(denoms []string) sdk.Coins {
	ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
	var coins sdk.Coins
	for _, denom := range denoms {
		createdDenom, err := suite.App.TokenFactoryKeeper.CreateDenom(ctx, suite.TestAccs[0].String(), denom)
		suite.Assert().Nil(err)
		coins = append(coins, sdk.NewCoin(createdDenom, sdkmath.NewInt(1000000000)))

	}
	err := suite.App.BankKeeper.MintCoins(ctx, types.ModuleName, coins)
	suite.Assert().Nil(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, suite.TestAccs[0], coins)
	suite.Assert().Nil(err)
	return coins
}

func (suite *IntegrationTestSuite) TestCreatePairMsg() {
	coins := suite.createTestToken([]string{"c1", "c2", "c3"})
	for _, tc := range []struct {
		desc                 string
		token0               string
		token1               string
		expecterr            error
		expectLpMetadataBase string
		expectPairId         uint64
	}{
		{
			desc:                 "normal",
			token0:               coins[0].Denom,
			token1:               coins[1].Denom,
			expecterr:            nil,
			expectLpMetadataBase: strings.Join([]string{types.ModuleLpPrefix, coins[0].Denom, coins[1].Denom}, "/"),
			expectPairId:         1,
		},
		{
			desc:                 "should sort tokens",
			token0:               coins[2].Denom,
			token1:               coins[1].Denom,
			expecterr:            nil,
			expectLpMetadataBase: strings.Join([]string{types.ModuleLpPrefix, coins[1].Denom, coins[2].Denom}, "/"),
			expectPairId:         2,
		},
		{
			desc:                 "create aleready existd pair should failed",
			token0:               coins[2].Denom,
			token1:               coins[1].Denom,
			expecterr:            types.ErrPairCreated,
			expectLpMetadataBase: strings.Join([]string{types.ModuleLpPrefix, coins[1].Denom, coins[2].Denom}, "/"),
			expectPairId:         0,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())
			_, err := suite.msgServer.CreatePair(ctx, &types.MsgCreatePair{
				Sender: suite.TestAccs[0].String(),
				Token0: tc.token0,
				Token1: tc.token1,
			})
			if tc.expecterr == nil {
				suite.Require().Nil(err)
				pair := suite.App.SwapKeeper.GetPairFromId(ctx, tc.expectPairId)
				sortToken0, sortToken1 := types.SortToken(tc.token0, tc.token1)
				suite.Require().Equal(pair.Token0.Denom, sortToken0)
				suite.Require().Equal(pair.Token1.Denom, sortToken1)
				lpmetadata, _ := suite.App.BankKeeper.GetDenomMetaData(ctx, pair.LpToken.Denom)
				suite.Require().Equal(lpmetadata.Base, pair.LpToken.Denom)
			} else {
				suite.Require().Equal(tc.expecterr, err)
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestAddLiquidityMsg() {
	coins := suite.createTestToken([]string{"a1", "a2", "c3"})
	ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())

	_, err := suite.msgServer.CreatePair(ctx, &types.MsgCreatePair{
		Sender: suite.TestAccs[0].String(),
		Token0: coins[0].Denom,
		Token1: coins[1].Denom,
	})

	suite.Require().Nil(err)

	for _, tc := range []struct {
		desc                string
		token0              sdk.Coin
		token1              sdk.Coin
		expecterr           error
		expectLpTokenAmount sdkmath.Int
	}{
		{
			desc:      "add to not exist pair",
			token0:    coins[0],
			token1:    coins[2],
			expecterr: types.ErrPairNotExist,
		},
		{
			desc:                "normal",
			token0:              sdk.NewCoin(coins[0].Denom, sdk.NewInt(1000000)),
			token1:              sdk.NewCoin(coins[1].Denom, sdk.NewInt(1000000)),
			expecterr:           nil,
			expectLpTokenAmount: sdk.NewInt(1000000),
		},
		{
			desc:                "add to stated pair",
			token0:              sdk.NewCoin(coins[0].Denom, sdk.NewInt(2000000)),
			token1:              sdk.NewCoin(coins[1].Denom, sdk.NewInt(2000000)),
			expecterr:           nil,
			expectLpTokenAmount: sdk.NewInt(2000000),
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			res, err := suite.msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
				Sender:     suite.TestAccs[0].String(),
				Token0:     tc.token0,
				Token1:     tc.token1,
				Amount0Min: sdkmath.ZeroInt(),
				Amount1Min: sdkmath.ZeroInt(),
				Deadline:   1000,
			})
			if tc.expecterr != nil {
				suite.Require().Equal(err, tc.expecterr)
			} else {
				suite.Require().Equal(tc.expectLpTokenAmount, res.LpToken.Amount)
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSwapExactTokensForTokensMsg() {
	coins := suite.createTestToken([]string{"se1", "se2", "se3", "se4"})
	ctx := suite.Ctx.WithEventManager(sdk.NewEventManager())

	for _, tc := range []struct {
		desc            string
		initPairs       []types.Pair
		expectedErr     error
		amountIn        sdk.Coin
		path            []string
		expectedOutCoin sdk.Coin
	}{
		{
			desc: "normal",
			initPairs: []types.Pair{
				{
					Token0: sdk.NewCoin(coins[0].Denom, sdk.NewInt(10000)),
					Token1: sdk.NewCoin(coins[1].Denom, sdk.NewInt(10000)),
				},
				{
					Token0: sdk.NewCoin(coins[1].Denom, sdk.NewInt(10000)),
					Token1: sdk.NewCoin(coins[2].Denom, sdk.NewInt(10000)),
				},
			},
			expectedErr:     nil,
			amountIn:        sdk.NewCoin(coins[0].Denom, sdk.NewInt(100)),
			path:            []string{coins[0].Denom, coins[1].Denom, coins[2].Denom},
			expectedOutCoin: sdk.NewCoin(coins[2].Denom, sdk.NewInt(96)),
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			for _, pair := range tc.initPairs {
				_, err := suite.msgServer.CreatePair(ctx, &types.MsgCreatePair{
					Sender: suite.TestAccs[0].String(),
					Token0: pair.Token0.Denom,
					Token1: pair.Token1.Denom,
				})

				suite.Require().Nil(err)

				_, err = suite.msgServer.AddLiquidity(ctx, &types.MsgAddLiquidity{
					Sender:     suite.TestAccs[0].String(),
					Token0:     pair.Token0,
					Token1:     pair.Token1,
					Amount0Min: sdkmath.ZeroInt(),
					Amount1Min: sdkmath.ZeroInt(),
					Deadline:   1000,
				})
				suite.Require().Nil(err)
			}

			res, err := suite.msgServer.SwapExactTokensForTokens(ctx, &types.MsgSwapExactTokensForTokens{
				Sender:       suite.TestAccs[0].String(),
				AmountIn:     tc.amountIn.Amount,
				AmountOutMin: sdk.ZeroInt(),
				Path:         tc.path,
				Recipient:    suite.TestAccs[0].String(),
				Deadline:     100,
			})

			if tc.expectedErr != nil {
				suite.Require().Nil(err)
				suite.Require().Equal(res.Ammounts[len(res.Ammounts)-1], tc.expectedOutCoin.Amount)
			}
		})
	}
}

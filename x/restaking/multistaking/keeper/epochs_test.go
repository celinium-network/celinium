package keeper_test

import (
	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RiseRateCalculateEquivalentCoin(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error) {
	return sdk.NewCoin(targetDenom, coin.Amount.QuoRaw(2)), nil
}

func DeclineRateCalculateEquivalentCoin(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error) {
	return sdk.NewCoin(targetDenom, coin.Amount.MulRaw(2)), nil
}

func (suite *KeeperTestSuite) TestRefreshDelegationAmountWhenRateRise() {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)

	err := suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})
	suite.Require().NoError(err)

	suite.app.MultiStakingKeeper.EquivalentCoinCalculator = RiseRateCalculateEquivalentCoin
	suite.app.MultiStakingKeeper.RefreshAgentDelegationAmount(suite.ctx)
}

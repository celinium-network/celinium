package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/app"
	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

var (
	PKs                     = simapp.CreateTestPubKeys(500)
	mockMultiRestakingDenom = "mmrd"
)

func createValAddrs(count int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := app.CreateIncrementalAccounts(count)
	valAddrs := app.ConvertAddrsToValAddrs(addrs)

	return addrs, valAddrs
}

// TODO Distinguish between integration tests and unit tests. it's look like integration tests?
func (suite *KeeperTestSuite) TestDelegate() {

	addrDels, _ := createValAddrs(1)
	vals := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))

	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)
	suite.mintCoin(multiRestakingCoin, addrDels[0])

	msg := types.MsgMultiStakingDelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: vals[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	}
	suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, msg)

	delegatorShares := suite.app.MultiStakingKeeper.GetMultiStakingShares(suite.ctx, 0, msg.DelegatorAddress)
	suite.Require().True(delegatorShares.Equal(multiRestakingCoin.Amount))

	agentID := suite.app.MultiStakingKeeper.GetLatestMultiStakingAgentID(suite.ctx)
	agent, found := suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))
	suite.Require().True(agent.Shares.Equal(delegatorShares))
}

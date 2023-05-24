package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

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
	delegatorAddrs, _ := createValAddrs(2)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))

	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])

	msg := types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	}
	suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, msg)

	agentID := suite.app.MultiStakingKeeper.GetLatestMultiStakingAgentID(suite.ctx)
	suite.Require().Equal(agentID, uint64(1))

	delegatorShares := suite.app.MultiStakingKeeper.GetMultiStakingShares(suite.ctx, agentID, msg.DelegatorAddress)
	suite.Require().True(delegatorShares.Equal(multiRestakingCoin.Amount))

	agent, found := suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))
	suite.Require().True(agent.Shares.Equal(delegatorShares))

	suite.mintCoin(multiRestakingCoin, delegatorAddrs[1])
	msg2 := types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[1].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	}
	suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, msg2)

	delegator2Shares := suite.app.MultiStakingKeeper.GetMultiStakingShares(suite.ctx, agentID, msg.DelegatorAddress)
	suite.Require().True(delegator2Shares.Equal(multiRestakingCoin.Amount))
	agent, found = suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount.MulRaw(2)))
	suite.Require().True(agent.Shares.Equal(delegatorShares.MulRaw(2)))
}

func (suite *KeeperTestSuite) TestUndelegate() {
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

	suite.app.MultiStakingKeeper.MultiStakingUndelegate(suite.ctx, &types.MsgMultiStakingUndelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})

	agentID := suite.app.MultiStakingKeeper.GetLatestMultiStakingAgentID(suite.ctx)

	delegator2Shares := suite.app.MultiStakingKeeper.GetMultiStakingShares(suite.ctx, agentID, delegatorAddrs[0].String())
	suite.Require().True(delegator2Shares.Equal(math.ZeroInt()))
	agent, found := suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(math.ZeroInt()))
	suite.Require().True(agent.Shares.Equal(math.ZeroInt()))

	// check unbonding records
	unbonding, found := suite.app.MultiStakingKeeper.GetMultiStakingUnbonding(suite.ctx, agentID, delegatorAddrs[0].String())
	suite.Require().True(found)
	suite.Require().Equal(len(unbonding.Entries), 1)

	entry := unbonding.Entries[0]
	suite.Require().True(entry.Balance.Equal(multiRestakingCoin))
	suite.Require().True(entry.InitialBalance.Equal(multiRestakingCoin))

	unbondingTime := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
	suite.Require().True(entry.CompletionTime.Equal(suite.ctx.BlockTime().Add(unbondingTime)))

	unbondingQueue := suite.app.MultiStakingKeeper.GetUBDQueueTimeSlice(suite.ctx, entry.CompletionTime)
	suite.Require().Equal(len(unbondingQueue), 1)
	unbondingDAPair := unbondingQueue[0]
	suite.Require().Equal(unbondingDAPair.AgentId, agentID)
	suite.Require().Equal(unbondingDAPair.DelegatorAddress, delegatorAddrs[0].String())
}

func (suite *KeeperTestSuite) TestUndelegateReward() {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)

	suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})

	rewardAmount := sdk.NewIntFromUint64(500000)
	rewardDenom := suite.app.StakingKeeper.GetParams(suite.ctx).BondDenom
	rewardCoins := sdk.Coins{sdk.NewCoin(rewardDenom, rewardAmount)}

	suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, rewardCoins)
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardCoins)

	agentID := suite.app.MultiStakingKeeper.GetLatestMultiStakingAgentID(suite.ctx)
	agent, _ := suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	valAddr, _ := sdk.ValAddressFromBech32(agent.ValidatorAddress)
	validator := suite.app.StakingKeeper.Validator(suite.ctx, valAddr)

	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, sdk.DecCoins{
		sdk.NewDecCoinFromDec(rewardDenom, sdk.NewDecFromInt(rewardAmount)),
	})

	suite.ctx = suite.ctx.
		WithBlockHeight(suite.ctx.BlockHeight() + 100).
		WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))

	suite.app.MultiStakingKeeper.MultiStakingUndelegate(suite.ctx, &types.MsgMultiStakingUndelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})

	// TODO check reward amount
}

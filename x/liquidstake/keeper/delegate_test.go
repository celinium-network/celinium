package keeper_test

import (
	"time"

	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestCreateNewDelegationRecordAtEpochStart() {
	suite.setSourceChainAndEpoch(suite.generateSourceChainParams(), suite.delegationEpoch())

	controlChainApp := getCeliniumApp(suite.controlChain)

	ctx := suite.controlChain.GetContext()
	nextDelegationRecordID := controlChainApp.LiquidStakeKeeper.GetDelegationRecordID(ctx)
	_, found := controlChainApp.LiquidStakeKeeper.GetDelegationRecord(ctx, nextDelegationRecordID-1)
	suite.True(found)
}

func (suite *KeeperTestSuite) TestUserDelegate() {
	sourceChainParams := suite.generateSourceChainParams()
	suite.setSourceChainAndEpoch(sourceChainParams, suite.delegationEpoch())

	ctlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	controlChainApp := getCeliniumApp(suite.controlChain)

	testCoin := suite.testCoin

	ctx := suite.controlChain.GetContext()
	bal := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalBefore := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.DerivativeDenom)

	err := controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, testCoin.Amount, ctlChainUserAddr)
	suite.NoError(err)

	balAfter := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalAfter := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.DerivativeDenom)

	suite.True(balAfter.Amount.Add(testCoin.Amount).Equal(bal.Amount))
	suite.True(derivativeBalAfter.Amount.Sub(testCoin.Amount).Equal(derivativeBalBefore.Amount))

	nextDelegationRecordID := controlChainApp.LiquidStakeKeeper.GetDelegationRecordID(ctx)
	delegationRecord, found := controlChainApp.LiquidStakeKeeper.GetDelegationRecord(ctx, nextDelegationRecordID-1)
	suite.True(found)
	suite.True(delegationRecord.DelegationCoin.Amount.Equal(testCoin.Amount))
}

func (suite *KeeperTestSuite) TestProcessDelegationAfterEpochAdvance() {
	sourceChainParams := suite.generateSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	// delegation at epoch 2
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	testCoin := suite.testCoin
	controlChainApp := getCeliniumApp(suite.controlChain)

	err := controlChainApp.LiquidStakeKeeper.Delegate(
		suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(epochInfo)

	sc, found := controlChainApp.LiquidStakeKeeper.GetSourceChain(suite.controlChain.GetContext(), sourceChainParams.ChainID)
	suite.True(found)
	suite.Equal(sc.StakedAmount, testCoin.Amount)
}

func (suite *KeeperTestSuite) delegationEpoch() *epochtypes.EpochInfo {
	return &epochtypes.EpochInfo{
		Identifier:              types.DelegationEpochIdentifier,
		StartTime:               suite.controlChain.CurrentHeader.Time,
		Duration:                time.Hour,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.controlChain.CurrentHeader.Time,
		EpochCountingStarted:    false,
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}
}

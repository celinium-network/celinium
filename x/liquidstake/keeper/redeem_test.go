package keeper_test

func (suite *KeeperTestSuite) TestRedeemAfterUnbondingComplete() {
	sourceChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	testCoin := suite.testCoin
	controlChainApp := getCeliniumApp(suite.controlChain)
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainUserAddr := ctlChainUserAccAddr.String()

	ctx := suite.controlChain.GetContext()
	err := controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(delegationEpochInfo)

	// user has already delegate, then undelegate
	unbondingEpochInfo := suite.unbondEpoch()
	ctx = suite.controlChain.GetContext()
	controlChainApp.EpochsKeeper.SetEpochInfo(ctx, *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	ctx = suite.controlChain.GetContext()
	controlChainApp.LiquidStakeKeeper.Undelegate(ctx, sourceChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)

	// process at next unbond epoch begin
	nextBlockTime := suite.advanceToNextEpoch(unbondingEpochInfo)
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)
	nextBlockWithRes(suite.sourceChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, ctlChainUserAddr)

	suite.WaitForUnbondingComplete(sourceChainParams, 2)

	ctx = suite.controlChain.GetContext()
	balBefore := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAccAddr, sourceChainParams.IbcDenom)
	err = controlChainApp.LiquidStakeKeeper.RedeemUndelegation(ctx, ctlChainUserAccAddr, 2, sourceChainParams.ChainID)
	balAfter := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAccAddr, sourceChainParams.IbcDenom)
	suite.NoError(err)
	suite.True(balAfter.Sub(balBefore).Amount.Equal(testCoin.Amount))
}

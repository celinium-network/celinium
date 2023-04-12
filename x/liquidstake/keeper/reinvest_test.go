package keeper_test

func (suite *KeeperTestSuite) TestReinvest() {
	// sourceChainParams := suite.mockSourceChainParams()
	// epochInfo := suite.delegationEpoch()
	// suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	// // delegation at epoch 2
	// sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	// controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	// testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	// mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	// suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	// controlChainApp := getCeliniumApp(suite.controlChain)

	// err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	// suite.NoError(err)

	// suite.processDelegation(epochInfo)

	// // finalize some block, so get staking reward.
	// stepDuration := (time.Hour)
	// step := 10
	// for i := 0; i < int(step-1); i++ {
	// 	suite.controlChain.Coordinator.CurrentTime = suite.controlChain.Coordinator.CurrentTime.Add(stepDuration)

	// 	fmt.Println(suite.sourceChain.Coordinator.CurrentTime.Format(time.RFC3339))

	// 	nextBlockWithRes(suite.controlChain, suite.controlChain.Coordinator.CurrentTime)
	// 	nextBlockWithRes(suite.sourceChain, suite.sourceChain.Coordinator.CurrentTime)

	// 	suite.transferPath.EndpointA.UpdateClient()
	// 	suite.transferPath.EndpointB.UpdateClient()
	// }
}

func (suite *KeeperTestSuite) TestSetWithdrawAddress() {
	sourceChainParams := suite.mockSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()

	withdrawOnSourceChain, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.UnboudAddress)
	suite.NoError(err)

	err = controlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	withdrawAccAddress := sourceChainApp.DistrKeeper.GetDelegatorWithdrawAddr(suite.sourceChain.GetContext(), withdrawOnSourceChain)
	withdrawOnSourceChain, err = controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.WithdrawAddress)

	suite.NoError(err)

	suite.Equal(withdrawAccAddress.String(), withdrawOnSourceChain.String())
}

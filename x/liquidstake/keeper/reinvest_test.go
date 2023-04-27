package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *KeeperTestSuite) TestReinvest() {
	srcChainParams := suite.generateSourceChainParams()
	delegationEpoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, delegationEpoch)

	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainUserAddr := ctlChainUserAccAddr.String()

	testCoin := suite.testCoin
	srcNativeDenom := srcChainParams.NativeDenom

	ctlChainApp := getCeliniumApp(suite.controlChain)
	srcChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()
	err := ctlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), ctlChainUserAddr)

	ctx = suite.controlChain.GetContext()
	_, err = ctlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(delegationEpoch)

	// set reward for validator in source chain.
	rewards := sdk.DecCoins{
		sdk.NewDecCoinFromDec(srcNativeDenom, sdk.NewDec(500000).Quo(sdk.NewDec(1))),
	}
	ctx = suite.sourceChain.GetContext()

	validatorNum := uint64(len(srcChainParams.Validators))
	rewardAmt := sdk.Coins{sdk.NewCoin(srcNativeDenom, sdk.NewIntFromUint64(50000*validatorNum))}

	srcChainApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, rewardAmt)
	srcChainApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardAmt)
	for _, v := range srcChainParams.Validators {
		valAddr, err := sdk.ValAddressFromBech32(v.Address)
		suite.NoError(err)
		valAcc := srcChainApp.StakingKeeper.Validator(ctx, valAddr)
		srcChainApp.DistrKeeper.AllocateTokensToValidator(ctx, valAcc, rewards)
	}

	// begin reinvest
	ctx = suite.controlChain.GetContext()
	ctlChainApp.LiquidStakeKeeper.StartReinvest(ctx)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), ctlChainUserAddr)

	ctx = suite.controlChain.GetContext()
	delegatorICAOnSrcChain, err := ctlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, srcChainParams.ConnectionID, srcChainParams.DelegateAddress)
	suite.NoError(err)

	// delegatorICAOnSrcChain has some reward now,
	ctx = suite.sourceChain.GetContext()
	balance := srcChainApp.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(delegatorICAOnSrcChain), srcNativeDenom)
	suite.True(balance.Amount.GT(sdk.ZeroInt()))

	// no user delegate. the reinvest the reward.
	suite.advanceEpochAndRelayIBC(suite.delegationEpoch())

	ctx = suite.controlChain.GetContext()
	srcChain, _ := ctlChainApp.LiquidStakeKeeper.GetSourceChain(ctx, srcChainParams.ChainID)
	// redeemrate has change
	suite.True((srcChain.Redemptionratio.GT(sdk.NewDec(1))))
	suite.True(srcChain.StakedAmount.Sub(testCoin.Amount).Sub(balance.Amount).Equal(sdk.ZeroInt()))
}

func (suite *KeeperTestSuite) TestSetWithdrawAddress() {
	sourceChainParams := suite.generateSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()

	delegatorAddr, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.DelegateAddress)
	suite.NoError(err)

	err = controlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	withdrawAccAddress := sourceChainApp.DistrKeeper.GetDelegatorWithdrawAddr(suite.sourceChain.GetContext(), sdk.MustAccAddressFromBech32(delegatorAddr))
	withdrawAddr, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.WithdrawAddress)

	suite.NoError(err)
	suite.Equal(withdrawAccAddress.String(), withdrawAddr)
}

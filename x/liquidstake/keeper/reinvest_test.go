package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *KeeperTestSuite) TestReinvest() {
	sourceChainParams := suite.generateSourceChainParams()
	delegationEpoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpoch)

	// delegation at epoch 2
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := suite.testCoin
	srcNativeDenom := sourceChainParams.NativeDenom

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()
	err := controlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	ctx = suite.controlChain.GetContext()
	err = controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(delegationEpoch)

	// set reward for validator in source chain.
	rewards := sdk.DecCoins{
		sdk.NewDecCoinFromDec(srcNativeDenom, sdk.NewDec(50000).Quo(sdk.NewDec(1))),
	}
	ctx = suite.sourceChain.GetContext()

	validatorNum := uint64(len(sourceChainParams.Validators))
	rewardAmt := sdk.Coins{sdk.NewCoin(srcNativeDenom, sdk.NewIntFromUint64(50000*validatorNum))}

	sourceChainApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, rewardAmt)
	sourceChainApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardAmt)
	for _, v := range sourceChainParams.Validators {
		valAddr, err := sdk.ValAddressFromBech32(v.Address)
		suite.NoError(err)
		valAcc := sourceChainApp.StakingKeeper.Validator(ctx, valAddr)
		sourceChainApp.DistrKeeper.AllocateTokensToValidator(ctx, valAcc, rewards)
	}

	// begin reinvest
	ctx = suite.controlChain.GetContext()
	controlChainApp.LiquidStakeKeeper.StartReInvest(ctx)
	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	ctx = suite.controlChain.GetContext()
	delegatorICAOnSrcChain, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.UnboudAddress)
	suite.NoError(err)

	// delegatorICAOnSrcChain has some reward now,
	ctx = suite.sourceChain.GetContext()
	balance := sourceChainApp.BankKeeper.GetBalance(ctx, delegatorICAOnSrcChain, srcNativeDenom)
	suite.False(balance.Amount.IsZero())

	// TODO where check reinvest effect?
}

func (suite *KeeperTestSuite) TestSetWithdrawAddress() {
	sourceChainParams := suite.generateSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()

	delegatorAddr, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.UnboudAddress)
	suite.NoError(err)

	err = controlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	withdrawAccAddress := sourceChainApp.DistrKeeper.GetDelegatorWithdrawAddr(suite.sourceChain.GetContext(), delegatorAddr)
	withdrawAddr, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		ctx, sourceChainParams.ConnectionID, sourceChainParams.WithdrawAddress)

	suite.NoError(err)
	suite.Equal(withdrawAccAddress.String(), withdrawAddr.String())
}

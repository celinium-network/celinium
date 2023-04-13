package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *KeeperTestSuite) TestReinvest() {
	sourceChainParams := suite.mockSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, epochInfo)

	// delegation at epoch 2
	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	ctx := suite.controlChain.GetContext()
	err := controlChainApp.LiquidStakeKeeper.SetDistriWithdrawAddress(ctx)
	suite.NoError(err)
	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

	err = controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceNextDelegationEpochAndProcess(epochInfo)

	// set reward for validator in sourcechain
	rewards := sdk.DecCoins{
		sdk.NewDecCoinFromDec(sourceChainParams.NativeDenom, sdk.NewDec(50000).Quo(sdk.NewDec(1))),
	}
	ctx = suite.sourceChain.GetContext()

	rewardAmt := sdk.Coins{sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(50000*uint64(len(sourceChainParams.Validators))))}
	sourceChainApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, rewardAmt)
	sourceChainApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardAmt)

	for _, v := range sourceChainParams.Validators {
		valAddr, err := sdk.ValAddressFromBech32(v.Address)
		suite.NoError(err)
		valAcc := sourceChainApp.StakingKeeper.Validator(ctx, valAddr)
		sourceChainApp.DistrKeeper.AllocateTokensToValidator(ctx, valAcc, rewards)
	}

	ctx = suite.controlChain.GetContext()
	controlChainApp.LiquidStakeKeeper.StartReInvest(ctx)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())
	withrawAddrOnSourceChain, err := controlChainApp.LiquidStakeKeeper.GetSourceChainAddr(
		suite.controlChain.GetContext(), sourceChainParams.ConnectionID, sourceChainParams.UnboudAddress)
	suite.NoError(err)

	balance := sourceChainApp.BankKeeper.GetBalance(suite.sourceChain.GetContext(), withrawAddrOnSourceChain, sourceChainParams.NativeDenom)
	suite.False(balance.Amount.IsZero())
}

func (suite *KeeperTestSuite) TestSetWithdrawAddress() {
	sourceChainParams := suite.mockSourceChainParams()
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

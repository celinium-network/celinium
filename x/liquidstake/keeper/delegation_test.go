package keeper_test

import (
	"math/big"
	"time"

	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/celinium-network/celinium/app/params"
)

func (suite *KeeperTestSuite) TestCreateNewProxyDelegationAtEpochStart() {
	suite.setSourceChainAndEpoch(suite.mockSourceChainParams(), suite.delegationEpoch())

	controlChainApp := getCeliniumApp(suite.controlChain)

	ctx := suite.controlChain.GetContext()
	nextProxyDelegationID := controlChainApp.LiquidStakeKeeper.GetProxyDelegationID(ctx)
	_, found := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, nextProxyDelegationID-1)
	suite.True(found)
}

func (suite *KeeperTestSuite) TestDelegate() {
	sourceChainParams := suite.mockSourceChainParams()
	suite.setSourceChainAndEpoch(sourceChainParams, suite.delegationEpoch())

	ctlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	controlChainApp := getCeliniumApp(suite.controlChain)

	testCoin := suite.testCoin

	ctx := suite.controlChain.GetContext()
	bal := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalBefore := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.DerivativeDenom)

	_, err := controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, testCoin.Amount, ctlChainUserAddr)
	suite.NoError(err)

	balAfter := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalAfter := controlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, sourceChainParams.DerivativeDenom)

	suite.True(balAfter.Amount.Add(testCoin.Amount).Equal(bal.Amount))
	suite.True(derivativeBalAfter.Amount.Sub(testCoin.Amount).Equal(derivativeBalBefore.Amount))

	nextProxyDelegationID := controlChainApp.LiquidStakeKeeper.GetProxyDelegationID(ctx)
	proxyDelegation, found := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, nextProxyDelegationID-1)
	suite.True(found)
	suite.True(proxyDelegation.Coin.Amount.Equal(testCoin.Amount))
}

func (suite *KeeperTestSuite) TestDelegateWithDiffRedeemRatio() {
	ratios := []sdk.Dec{
		sdk.NewDecWithPrec(111111, 5), // 1.11111
		sdk.NewDecWithPrec(99999, 5),  // 0.99999
	}

	srcChainParams := suite.mockSourceChainParams()
	suite.setSourceChainAndEpoch(srcChainParams, suite.delegationEpoch())

	ctlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainApp := getCeliniumApp(suite.controlChain)
	ctx := suite.controlChain.GetContext()
	testCoin := suite.testCoin

	delegateAmont := sdk.NewIntFromBigInt(big.NewInt(0).Div(testCoin.Amount.BigInt(), big.NewInt(int64(len(ratios)))))

	for _, ratio := range ratios {
		srcChainParams.Redemptionratio = ratio
		ctlChainApp.LiquidStakeKeeper.SetSourceChain(ctx, srcChainParams)

		derivativeBalBefore := ctlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, srcChainParams.DerivativeDenom)

		_, err := ctlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID,
			delegateAmont, ctlChainUserAddr)

		suite.NoError(err)
		derivativeBalAfter := ctlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAddr, srcChainParams.DerivativeDenom)
		derivativeAmt := derivativeBalAfter.Amount.Sub(derivativeBalBefore.Amount)
		suite.True(derivativeAmt.Equal(sdk.NewDecFromInt(delegateAmont).Quo(ratio).TruncateInt()))
	}
}

func (suite *KeeperTestSuite) TestProcessDelegationAfterAdvanceEpoch() {
	srcChainParams := suite.mockSourceChainParams()
	epochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, epochInfo)

	// delegation at epoch 2
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	controlChainApp := getCeliniumApp(suite.controlChain)
	testCoin := suite.testCoin

	ctx := suite.controlChain.GetContext()
	_, err := controlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(epochInfo)

	ctx = suite.controlChain.GetContext()
	sc, found := controlChainApp.LiquidStakeKeeper.GetSourceChain(ctx, srcChainParams.ChainID)
	suite.True(found)
	suite.Equal(sc.StakedAmount, testCoin.Amount)
}

func (suite *KeeperTestSuite) delegationEpoch() *epochtypes.EpochInfo {
	return &epochtypes.EpochInfo{
		Identifier:              appparams.DelegationEpochIdentifier,
		StartTime:               suite.controlChain.CurrentHeader.Time,
		Duration:                time.Hour,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.controlChain.CurrentHeader.Time,
		EpochCountingStarted:    false,
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}
}

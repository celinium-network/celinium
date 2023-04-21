package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	"github.com/celinium-network/celinium/app"
	appparams "github.com/celinium-network/celinium/app/params"
	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	"github.com/celinium-network/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) unbondEpoch() *epochtypes.EpochInfo {
	return &epochtypes.EpochInfo{
		Identifier:              appparams.UndelegationEpochIdentifier,
		StartTime:               suite.controlChain.CurrentHeader.Time,
		Duration:                time.Hour * 24,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.controlChain.CurrentHeader.Time,
		EpochCountingStarted:    true,
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}
}

func (suite *KeeperTestSuite) TestCreateEpochUnbonding() {
	ctlChainApp := getCeliniumApp(suite.controlChain)
	ctx := suite.controlChain.GetContext()
	ctlChainApp.EpochsKeeper.SetEpochInfo(ctx, *suite.delegationEpoch())

	suite.setSourceChainAndEpoch(suite.generateSourceChainParams(), suite.unbondEpoch())

	// check epoch unbonding at epoch 2
	controlChainApp := getCeliniumApp(suite.controlChain)
	unbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)

	suite.True(found)
	suite.Equal(len(unbonding.Unbondings), 0)
}

func (suite *KeeperTestSuite) TestUndelegate() {
	srcChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, delegationEpochInfo)

	testCoin := suite.testCoin
	ctlChainApp := getCeliniumApp(suite.controlChain)
	srcChainApp := getCeliniumApp(suite.sourceChain)

	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainUserAddr := ctlChainUserAccAddr.String()

	ctx := suite.controlChain.GetContext()
	_, err := ctlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)
	suite.NoError(err)

	suite.advanceEpochAndRelayIBC(delegationEpochInfo)

	// user has already delegate, then undelegate
	unbondingEpochInfo := suite.unbondEpoch()
	ctx = suite.controlChain.GetContext()
	ctlChainApp.EpochsKeeper.SetEpochInfo(ctx, *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	ctx = suite.controlChain.GetContext()
	derivativeBalBefore := ctlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAccAddr, srcChainParams.DerivativeDenom)
	ctlChainApp.LiquidStakeKeeper.Undelegate(ctx, srcChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)

	derivativeBalAfter := ctlChainApp.BankKeeper.GetBalance(ctx, ctlChainUserAccAddr, srcChainParams.DerivativeDenom)
	suite.True(derivativeBalBefore.Sub(derivativeBalAfter).Amount.Equal(testCoin.Amount))

	// process at next unbond epoch begin
	nextBlockTime := suite.advanceToNextEpoch(unbondingEpochInfo)
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)
	nextBlockWithRes(suite.sourceChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, ctlChainUserAddr)

	ctx = suite.controlChain.GetContext()
	epochUnbonding, found := ctlChainApp.LiquidStakeKeeper.GetEpochUnboundings(ctx, 2)
	suite.True(found)
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != srcChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, types.UnbondingWaitting)

		suite.True(unbonding.BurnedDerivativeAmount.Equal(testCoin.Amount))
		suite.True(unbonding.RedeemNativeToken.Amount.Equal(testCoin.Amount))
	}

	// check unbonding in source chain
	delegatorOnSourceChain, _ := ctlChainApp.LiquidStakeKeeper.GetSourceChainAddr(ctx, srcChainParams.ConnectionID, srcChainParams.DelegateAddress)

	ctx = suite.sourceChain.GetContext()
	UnbondingOnSrcChain := srcChainApp.StakingKeeper.GetAllUnbondingDelegations(ctx, sdk.MustAccAddressFromBech32(delegatorOnSourceChain))
	allocatedUnbonding := srcChainParams.AllocateFundsForValidator(testCoin.Amount)
	for _, unbonding := range UnbondingOnSrcChain {
		for _, alloc := range allocatedUnbonding {
			if alloc.Address != unbonding.ValidatorAddress {
				continue
			}
			totalUnbondingBal := math.ZeroInt()
			for _, e := range unbonding.Entries {
				totalUnbondingBal = totalUnbondingBal.Add(e.InitialBalance)
			}
			suite.True(alloc.Amount.Equal(totalUnbondingBal))
		}
	}

	ctx = suite.controlChain.GetContext()
	sourceChainAfter, _ := ctlChainApp.LiquidStakeKeeper.GetSourceChain(ctx, srcChainParams.ChainID)
	suite.True(sourceChainAfter.StakedAmount.Equal(sdk.ZeroInt()))
}

func (suite *KeeperTestSuite) TestWithdrawCompleteUnbond() {
	sourceChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	testCoin := suite.testCoin
	controlChainApp := getCeliniumApp(suite.controlChain)
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainUserAddr := ctlChainUserAccAddr.String()

	ctx := suite.controlChain.GetContext()
	_, err := controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, testCoin.Amount, ctlChainUserAccAddr)
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

	epochUnbonding, _ := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, types.UnbondingDone)
	}
	amt := controlChainApp.BankKeeper.GetBalance(
		suite.controlChain.GetContext(),
		sdk.MustAccAddressFromBech32(sourceChainParams.DelegateAddress),
		sourceChainParams.IbcDenom,
	)

	suite.False(amt.Amount.IsZero())
}

func (suite *KeeperTestSuite) WaitForUnbondingComplete(sourceChainParams *types.SourceChain, unbondingEpoch uint64) {
	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	ctlChainUserAddr := ctlChainUserAccAddr.String()

	ctx := suite.controlChain.GetContext()
	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(ctx, unbondingEpoch)
	suite.True(found)

	var unbondCompleteTime time.Time
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, types.UnbondingWaitting)
		unbondCompleteTime = time.Unix(0, int64(unbonding.UnbondTIme))
	}

	// make the light client not expired.
	stepDuration := (time.Hour * 24)
	step := app.DefaultUnbondingTime / (time.Hour * 24)
	for i := 0; i < int(step-1); i++ {
		suite.controlChain.Coordinator.CurrentTime = suite.controlChain.Coordinator.CurrentTime.Add(stepDuration)
		nextBlockWithRes(suite.controlChain, suite.controlChain.Coordinator.CurrentTime)
		nextBlockWithRes(suite.sourceChain, suite.sourceChain.Coordinator.CurrentTime)

		suite.transferPath.EndpointA.UpdateClient()
		suite.transferPath.EndpointB.UpdateClient()
	}

	unbondCompleteTime = unbondCompleteTime.Add(time.Minute * 20)
	nextBlockWithRes(suite.sourceChain, unbondCompleteTime)
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, unbondCompleteTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.controlChain.Coordinator.CurrentTime = unbondCompleteTime

	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, nextBlockBeginRes.Events, ctlChainUserAddr)
	ctx = suite.sourceChain.GetContext()
	sourceChainApp.IBCKeeper.RecvPacket(ctx, &msgRecvPackets[0])

	suite.sourceChain.NextBlock()
	suite.transferPath.EndpointB.UpdateClient()

	events := ctx.EventManager().Events()
	ack, _ := assembleAckPacketFromEvents(suite.sourceChain, msgRecvPackets[0].Packet, events)
	recvs := parseMsgRecvPacketFromEvents(suite.sourceChain, events.ToABCIEvents(), ctlChainUserAddr)

	for i := 0; i < len(recvs); i++ {
		controlChainApp.IBCKeeper.RecvPacket(suite.controlChain.GetContext(), &recvs[i])
	}
	controlChainApp.IBCKeeper.Acknowledgement(suite.controlChain.GetContext(), ack)
}

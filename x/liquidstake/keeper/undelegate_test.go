package keeper_test

import (
	"time"

	"github.com/celinium-netwok/celinium/app"
	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) unbondEpoch() *epochtypes.EpochInfo {
	return &epochtypes.EpochInfo{
		Identifier:              types.UndelegationEpochIdentifier,
		StartTime:               suite.controlChain.CurrentHeader.Time,
		Duration:                time.Hour * 24,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.controlChain.CurrentHeader.Time,
		EpochCountingStarted:    true,
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}
}

func (suite *KeeperTestSuite) TestCreateEpochUnbonding() {
	suite.setSourceChainAndEpoch(suite.generateSourceChainParams(), suite.unbondEpoch())

	// check epoch unbonding at epoch 2
	controlChainApp := getCeliniumApp(suite.controlChain)
	unbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)

	suite.True(found)
	suite.Equal(len(unbonding.Unbondings), 0)
}

func (suite *KeeperTestSuite) TestUserUndelegate() {
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

	ctx = suite.controlChain.GetContext()
	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(ctx, 2)
	suite.True(found)
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, uint32(types.UnbondingWaitting))

		// TODO check the unbonding.Amount will be failed
	}
}

func (suite *KeeperTestSuite) advanceToNextEpoch(epochInfo *epochtypes.EpochInfo) time.Time {
	coordTime := suite.controlChain.Coordinator.CurrentTime
	duration := epochInfo.Duration - (coordTime.Sub(epochInfo.StartTime.Add(epochInfo.Duration)))
	coordTime = coordTime.Add(duration + time.Minute*5)

	suite.controlChain.Coordinator.CurrentTime = coordTime
	suite.sourceChain.Coordinator.CurrentTime = coordTime

	return coordTime
}

func (suite *KeeperTestSuite) TestWithdrawCompleteUnbond() {
	sourceChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	testCoin := suite.testCoin
	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)
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

	ctx = suite.controlChain.GetContext()
	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(ctx, 2)
	suite.True(found)

	var unbondCompleteTime time.Time
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, uint32(types.UnbondingWaitting))
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
	_, nextBlockBeginRes = nextBlockWithRes(suite.controlChain, unbondCompleteTime)

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

	epochUnbonding, _ = controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		// TODO should be UndelegationCliamble
		suite.Equal(unbonding.Status, uint32(types.UndelegationComplete))
	}
	amt := controlChainApp.BankKeeper.GetBalance(
		suite.controlChain.GetContext(),
		sdk.MustAccAddressFromBech32(sourceChainParams.UnboudAddress),
		sourceChainParams.IbcDenom,
	)

	// TODO more check
	suite.False(amt.Amount.IsZero())
}

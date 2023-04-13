package keeper_test

import (
	"fmt"
	"time"

	"github.com/celinium-netwok/celinium/app"
	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) unbondingEpoch() *epochtypes.EpochInfo {
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
	suite.setSourceChainAndEpoch(suite.mockSourceChainParams(), suite.unbondingEpoch())

	// check epoch unbonding at epoch 2
	controlChainApp := getCeliniumApp(suite.controlChain)
	unbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
	suite.True(found)

	suite.Equal(len(unbonding.Unbondings), 0)
}

func (suite *KeeperTestSuite) TestUserUndelegate() {
	sourceChainParams := suite.mockSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	controlChainApp := getCeliniumApp(suite.controlChain)

	err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceNextDelegationEpochAndProcess(delegationEpochInfo)

	// user has already delegate, then undelegate
	unbondingEpochInfo := suite.unbondingEpoch()
	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	controlChainApp.LiquidStakeKeeper.Undelegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)

	// process at next unbond epoch begin
	coordTime := suite.advanceToNextEpoch(unbondingEpochInfo)

	nextBlockTime := coordTime
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)
	nextBlockWithRes(suite.sourceChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, controlChainUserAddr.String())
	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
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
	sourceChainParams := suite.mockSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.advanceNextDelegationEpochAndProcess(delegationEpochInfo)

	// user has already delegate, then undelegate
	unbondingEpochInfo := suite.unbondingEpoch()
	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	controlChainApp.LiquidStakeKeeper.Undelegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)

	// process at next unbond epoch begin
	coordTime := suite.advanceToNextEpoch(unbondingEpochInfo)

	nextBlockTime := coordTime
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)
	nextBlockWithRes(suite.sourceChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, controlChainUserAddr.String())

	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
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

		fmt.Println(suite.sourceChain.Coordinator.CurrentTime.Format(time.RFC3339))

		nextBlockWithRes(suite.controlChain, suite.controlChain.Coordinator.CurrentTime)
		nextBlockWithRes(suite.sourceChain, suite.sourceChain.Coordinator.CurrentTime)

		suite.transferPath.EndpointA.UpdateClient()
		suite.transferPath.EndpointB.UpdateClient()
	}

	unbondCompleteTime = unbondCompleteTime.Add(time.Minute * 20)

	nextBlockWithRes(suite.sourceChain, unbondCompleteTime)
	fmt.Println(unbondCompleteTime.Format(time.RFC3339))
	fmt.Println(suite.sourceChain.Coordinator.CurrentTime.Format(time.RFC3339))
	_, nextBlockBeginRes = nextBlockWithRes(suite.controlChain, unbondCompleteTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.controlChain.Coordinator.CurrentTime = unbondCompleteTime

	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, nextBlockBeginRes.Events, controlChainUserAddr.String())
	ctx := suite.sourceChain.GetContext()
	sourceChainApp.IBCKeeper.RecvPacket(ctx, &msgRecvPackets[0])

	suite.sourceChain.NextBlock()
	suite.transferPath.EndpointB.UpdateClient()

	ack, _ := assembleAckPacketFromEvents(suite.sourceChain, msgRecvPackets[0].Packet, ctx.EventManager().Events())
	recvs := parseMsgRecvPacketFromEvents(suite.sourceChain, ctx.EventManager().ABCIEvents(), controlChainUserAddr.String())

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

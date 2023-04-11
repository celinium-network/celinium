package keeper_test

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
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

	suite.processDelegation(delegationEpochInfo)

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

func (suite *KeeperTestSuite) relayIBCPacketFromCtlToSrc(events []abci.Event, sender string) {
	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, events, sender)

	for i := 0; i < len(msgRecvPackets); i++ {
		var midRecvMsg *channeltypes.MsgRecvPacket
		var midAckMsg *channeltypes.MsgAcknowledgement

		midAckMsg, err := chainRecvPacket(suite.sourceChain, suite.transferPath.EndpointB, &msgRecvPackets[i])

		if midAckMsg == nil || err != nil {
			break
		}

		// relay the ibc msg unitl no ibc info in events.
		for {
			midRecvMsg, _ = chainRecvAck(suite.controlChain, suite.transferPath.EndpointA, midAckMsg)
			if midRecvMsg == nil {
				break
			}

			midAckMsg, _ = chainRecvPacket(suite.sourceChain, suite.transferPath.EndpointB, midRecvMsg)
			if midAckMsg == nil {
				break
			}
		}
	}
}

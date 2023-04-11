package keeper_test

import (
	"time"

	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestCreateNewDelegationRecordAtEpochStart() {
	sourceChain := suite.mockSourceChainParams()
	suite.startDelegationEpoch(sourceChain)

	controlChainApp := getCeliniumApp(suite.controlChain)

	// check new delegation record
	nextDelegationRecordID := controlChainApp.LiquidStakeKeeper.GetDelegationRecordID(suite.controlChain.GetContext())
	_, found := controlChainApp.LiquidStakeKeeper.GetDelegationRecord(suite.controlChain.GetContext(), nextDelegationRecordID-1)
	suite.True(found)
}

func (suite *KeeperTestSuite) TestUserDelegate() {
	sourceChainParams := suite.mockSourceChainParams()
	suite.startDelegationEpoch(sourceChainParams)

	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	controlChainApp := getCeliniumApp(suite.controlChain)

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	ibcBalanceBeforeDelegate := controlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalanceBeforeDelegate := controlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, sourceChainParams.DerivativeDenom)

	err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	ibcBalanceAfterDelegate := controlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, sourceChainParams.IbcDenom)
	derivativeBalanceAfterDelegate := controlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, sourceChainParams.DerivativeDenom)

	suite.True(ibcBalanceAfterDelegate.Amount.Add(testCoin.Amount).Equal(ibcBalanceBeforeDelegate.Amount))
	suite.True(derivativeBalanceAfterDelegate.Amount.Sub(testCoin.Amount).Equal(derivativeBalanceBeforeDelegate.Amount))

	nextDelegationRecordID := controlChainApp.LiquidStakeKeeper.GetDelegationRecordID(suite.controlChain.GetContext())
	delegationRecord, found := controlChainApp.LiquidStakeKeeper.GetDelegationRecord(suite.controlChain.GetContext(), nextDelegationRecordID-1)
	suite.True(found)
	suite.True(delegationRecord.DelegationCoin.Amount.Equal(testCoin.Amount))
}

func (suite *KeeperTestSuite) TestProcessDelegation() {
	sourceChainParams := suite.mockSourceChainParams()
	suite.startDelegationEpoch(sourceChainParams)

	// delegation at epoch 2
	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	controlChainApp := getCeliniumApp(suite.controlChain)

	err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	epochInfo, found := controlChainApp.EpochsKeeper.GetEpochInfo(suite.controlChain.GetContext(), types.DelegationEpochIdentifier)
	suite.True(found)

	coordTime := suite.controlChain.Coordinator.CurrentTime
	duration := time.Hour - (coordTime.Sub(epochInfo.StartTime.Add(time.Hour)))

	// make next block will start new delegation epoch
	coordTime = coordTime.Add(duration + time.Minute*5)

	suite.controlChain.Coordinator.CurrentTime = coordTime
	suite.sourceChain.Coordinator.CurrentTime = coordTime

	nextBlockTime := coordTime
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, nextBlockBeginRes.Events, controlChainUserAddr.String())

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

	// delegate successful
	sc, found := controlChainApp.LiquidStakeKeeper.GetSourceChain(suite.controlChain.GetContext(), sourceChainParams.ChainID)
	suite.True(found)
	suite.Equal(sc.StakedAmount, testCoin.Amount)
}

func (suite *KeeperTestSuite) startDelegationEpoch(sourceChain *types.SourceChain) *epochtypes.EpochInfo {
	controlChainApp := getCeliniumApp(suite.controlChain)

	channelSequence := controlChainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := controlChainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), sourceChain)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	createdICAs := getCreatedICAFromSourceChain(sourceChain)
	for _, ica := range createdICAs {
		suite.relayICACreatedPacket(channelSequence, ica)
		channelSequence++
	}

	// set delegation Epoch Info
	epochInfo := epochtypes.EpochInfo{
		Identifier:              types.DelegationEpochIdentifier,
		StartTime:               suite.controlChain.CurrentHeader.Time,
		Duration:                time.Hour,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.controlChain.CurrentHeader.Time,
		EpochCountingStarted:    true,
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}

	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), epochInfo)

	// start epoch and update off chain light.
	suite.controlChain.Coordinator.IncrementTimeBy(time.Hour)
	suite.transferPath.EndpointA.UpdateClient()

	return &epochInfo
}

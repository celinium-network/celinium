package keeper_test

import (
	"time"

	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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
	sourceChainApp := getCeliniumApp(suite.sourceChain)

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
	_, nextBlockBeginRes := NextBlockWithRes(suite.controlChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, nextBlockBeginRes.Events, controlChainUserAddr.String())

	for i := 0; i < len(msgRecvPackets); i++ {
		sourceChainContext := suite.sourceChain.GetContext()
		_, err = sourceChainApp.IBCKeeper.RecvPacket(sourceChainContext, &msgRecvPackets[i])
		suite.NoError(err)
		suite.sourceChain.NextBlock()
		suite.transferPath.EndpointB.UpdateClient()

		ackMsg, err := assembleAckPacketFromEvents(suite.sourceChain, msgRecvPackets[i].Packet, sourceChainContext.EventManager().Events())
		suite.NoError(err)

		controlChainContext := suite.controlChain.GetContext()
		_, err = controlChainApp.IBCKeeper.Acknowledgement(controlChainContext, ackMsg)
		suite.NoError(err)
		suite.controlChain.NextBlock()
		suite.transferPath.EndpointA.UpdateClient()

		recvMsg, err := assembleRecvPacketByEvents(suite.controlChain, controlChainContext.EventManager().Events())
		suite.NoError(err)

		sourceChainContext = suite.sourceChain.GetContext()
		_, err = sourceChainApp.IBCKeeper.RecvPacket(sourceChainContext, recvMsg)
		suite.NoError(err)
		suite.sourceChain.NextBlock()
		suite.transferPath.EndpointB.UpdateClient()

		ackMsg, err = assembleAckPacketFromEvents(suite.sourceChain, recvMsg.Packet, sourceChainContext.EventManager().Events())
		suite.NoError(err)

		controlChainContext = suite.controlChain.GetContext()
		_, err = controlChainApp.IBCKeeper.Acknowledgement(controlChainContext, ackMsg)
		suite.NoError(err)
	}

	// delegate successful
	sc, found := controlChainApp.LiquidStakeKeeper.GetSourceChain(suite.controlChain.GetContext(), sourceChainParams.ChainID)
	suite.True(found)
	suite.Equal(sc.StakedAmount, testCoin.Amount)
}

func assembleRecvPacketByEvents(chain *ibctesting.TestChain, events sdk.Events) (*channeltypes.MsgRecvPacket, error) {
	packet, err := ibctesting.ParsePacketFromEvents(events)
	if err != nil {
		return nil, err
	}

	commitKey := ibchost.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
	proof, height := chain.QueryProof(commitKey)

	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)

	msgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          packet,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          chain.SenderAccount.GetAddress().String(),
	}

	return &msgRecvPacket, nil
}

func assembleAckPacketFromEvents(chain *ibctesting.TestChain, packet channeltypes.Packet, events sdk.Events) (*channeltypes.MsgAcknowledgement, error) {
	ack, err := ibctesting.ParseAckFromEvents(events)
	if err != nil {
		return nil, err
	}
	key := ibchost.PacketAcknowledgementKey(packet.GetDestPort(),
		packet.GetDestChannel(),
		packet.GetSequence())

	proof, height := chain.QueryProof(key)

	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)

	ackMsg := channeltypes.MsgAcknowledgement{
		Packet:          packet,
		Acknowledgement: ack,
		ProofAcked:      proof,
		ProofHeight:     height,
		Signer:          chain.SenderAccount.GetAddress().String(),
	}

	return &ackMsg, nil
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

func parseMsgRecvPacketFromEvents(fromChain *ibctesting.TestChain, events []abci.Event, sender string) []channeltypes.MsgRecvPacket {
	var msgRecvPackets []channeltypes.MsgRecvPacket
	for _, ev := range events {
		events := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(events)
		if err != nil {
			continue
		}

		commitKey := ibchost.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		proof, height := fromChain.QueryProof(commitKey)

		backProofType := ibccommitmenttypes.MerkleProof{}
		backProofType.Unmarshal(proof)

		msgRecvPacket := channeltypes.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          sender,
		}

		msgRecvPackets = append(msgRecvPackets, msgRecvPacket)
	}

	return msgRecvPackets
}

// next block with res
func NextBlockWithRes(chain *ibctesting.TestChain, nextBlockTime time.Time) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
	endBlockres := chain.App.EndBlock(abci.RequestEndBlock{Height: chain.CurrentHeader.Height})

	chain.App.Commit()

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain.T, chain.Vals, endBlockres.ValidatorUpdates)

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.CurrentHeader.ProposerAddress,
	}

	chain.CurrentHeader.Time = nextBlockTime
	beginBlockRes := chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader})

	chain.NextBlock()
	return endBlockres, beginBlockRes
}

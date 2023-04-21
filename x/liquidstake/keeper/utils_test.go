package keeper_test

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	"github.com/celinium-network/celinium/app"
	"github.com/celinium-network/celinium/app/params"
	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestIBCTransfer() {
	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	ctlChainApp := getCeliniumApp(suite.controlChain)

	coin := sdk.NewCoin(params.DefaultBondDenom, sdk.NewIntFromUint64(1000000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, coin)

	ibcDenom := suite.calcuateIBCDenom(
		suite.transferPath.EndpointB.ChannelConfig.PortID,
		suite.transferPath.EndpointB.ChannelID,
		params.DefaultBondDenom)

	ctlBalanceBefore := ctlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, ibcDenom)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), coin, suite.transferPath, true)

	ctlBalanceAfter := ctlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, ibcDenom)

	suite.Equal(ctlBalanceAfter.Amount.Sub(ctlBalanceBefore.Amount), coin.Amount)
}

func (suite *KeeperTestSuite) IBCTransfer(
	from string,
	to string,
	coin sdk.Coin,
	transferpath *ibctesting.Path,
	transferForward bool,
) {
	srcEndpoint := transferpath.EndpointA
	destEndpoint := transferpath.EndpointB
	if !transferForward {
		srcEndpoint, destEndpoint = destEndpoint, srcEndpoint
	}

	destChainApp := getCeliniumApp(destEndpoint.Chain)

	timesout := srcEndpoint.Chain.CurrentHeader.Time.Add(time.Hour * 5).UnixNano()
	msg := ibctransfertypes.NewMsgTransfer(
		srcEndpoint.ChannelConfig.PortID,
		srcEndpoint.ChannelID,
		coin,
		from,
		to,
		ibcclienttypes.Height{},
		uint64(timesout),
		"",
	)

	res, err := srcEndpoint.Chain.SendMsgs(msg)
	suite.NoError(err)

	suite.Require().NoError(err)
	suite.transferPath.EndpointB.UpdateClient()

	for _, ev := range res.Events {
		events := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(events)
		if err != nil {
			continue
		}

		commitKey := ibchost.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		proof, height := srcEndpoint.Chain.QueryProof(commitKey)

		msgRecvPacket := channeltypes.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          from,
		}

		_, err = destChainApp.IBCKeeper.RecvPacket(destEndpoint.Chain.GetContext(), &msgRecvPacket)
		suite.NoError(err)
	}
}

func (suite *KeeperTestSuite) setSourceChainAndEpoch(sourceChain *types.SourceChain, epochInfo *epochtypes.EpochInfo) {
	controlChainApp := getCeliniumApp(suite.controlChain)

	// start epoch and update offchain light.
	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), *epochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(epochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	suite.setSourceChain(controlChainApp, sourceChain)
}

func (suite *KeeperTestSuite) setSourceChain(chainApp *app.App, sourceChain *types.SourceChain) {
	channelSequence := chainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := chainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), sourceChain)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	createdICAs := getCreatedICAFromSourceChain(sourceChain)
	for _, ica := range createdICAs {
		suite.relayICACreatedPacket(channelSequence, ica)
		channelSequence++
	}
}

func mintCoin(
	chain *ibctesting.TestChain,
	to sdk.AccAddress,
	coin sdk.Coin,
) {
	chainApp := getCeliniumApp(chain)

	chainApp.BankKeeper.MintCoins(
		chain.GetContext(),
		ibctransfertypes.ModuleName,
		sdk.NewCoins(coin),
	)
	chainApp.BankKeeper.SendCoinsFromModuleToAccount(
		chain.GetContext(),
		ibctransfertypes.ModuleName,
		to,
		sdk.NewCoins(coin))
}

func chainRecvPacket(chain *ibctesting.TestChain, endpoint *ibctesting.Endpoint, msgRecvpacket *channeltypes.MsgRecvPacket) (*channeltypes.MsgAcknowledgement, error) {
	ctx := chain.GetContext()
	getCeliniumApp(chain)
	if _, err := getCeliniumApp(chain).IBCKeeper.RecvPacket(ctx, msgRecvpacket); err != nil {
		return nil, err
	}

	chain.NextBlock()
	endpoint.UpdateClient()

	return assembleAckPacketFromEvents(chain, msgRecvpacket.Packet, ctx.EventManager().Events())
}

func chainRecvAck(chain *ibctesting.TestChain, endpoint *ibctesting.Endpoint, ack *channeltypes.MsgAcknowledgement) (*channeltypes.MsgRecvPacket, error) {
	ctx := chain.GetContext()
	getCeliniumApp(chain)
	if _, err := getCeliniumApp(chain).IBCKeeper.Acknowledgement(ctx, ack); err != nil {
		return nil, err
	}

	chain.NextBlock()
	endpoint.UpdateClient()

	return assembleRecvPacketByEvents(chain, ctx.EventManager().Events())
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

func parseMsgRecvPacketFromEvents(fromChain *ibctesting.TestChain, events []abci.Event, sender string) []channeltypes.MsgRecvPacket {
	var msgRecvPackets []channeltypes.MsgRecvPacket
	for _, ev := range events {
		sdkevents := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(sdkevents)
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

func (suite *KeeperTestSuite) advanceEpochAndRelayIBC(epochInfo *epochtypes.EpochInfo) {
	ctlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()
	coordTime := suite.controlChain.Coordinator.CurrentTime
	duration := epochInfo.Duration - coordTime.Sub(epochInfo.StartTime.Add(epochInfo.Duration))

	// make next block will start new epoch
	coordTime = coordTime.Add(duration + time.Minute*5)

	suite.controlChain.Coordinator.CurrentTime = coordTime
	suite.sourceChain.Coordinator.CurrentTime = coordTime

	nextBlockTime := coordTime
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, ctlChainUserAddr.String())
}

func (suite *KeeperTestSuite) relayIBCPacketFromCtlToSrc(events []abci.Event, sender string) {
	msgRecvPackets := parseMsgRecvPacketFromEvents(suite.controlChain, events, sender)

	channelPackets := make(map[string][]channeltypes.MsgRecvPacket)
	channelProcesed := make(map[string]int)

	for _, packet := range msgRecvPackets {
		ps, ok := channelPackets[packet.Packet.SourceChannel]
		if !ok {
			ps = make([]channeltypes.MsgRecvPacket, 0)
		}
		ps = append(ps, packet)

		channelPackets[packet.Packet.SourceChannel] = ps
		channelProcesed[packet.Packet.SourceChannel] = 0
	}

	for {
		for channelID, packets := range channelPackets {
			offset := channelProcesed[channelID]
			for ; offset < len(packets); offset++ {

				midAckMsg, _ := chainRecvPacket(suite.sourceChain, suite.transferPath.EndpointB, &packets[offset])
				if midAckMsg == nil {
					continue
				}

				midRecvMsg, _ := chainRecvAck(suite.controlChain, suite.transferPath.EndpointA, midAckMsg)
				if midRecvMsg == nil {
					continue
				}
				if midRecvMsg.Packet.SourceChannel == channelID {
					packets = append(packets, *midRecvMsg)
				} else {
					ps, ok := channelPackets[midRecvMsg.Packet.SourceChannel]
					if !ok {
						ps = make([]channeltypes.MsgRecvPacket, 0)
					}
					ps = append(ps, *midRecvMsg)
					channelPackets[midRecvMsg.Packet.SourceChannel] = ps
				}
			}
			channelProcesed[channelID] = offset
		}

		done := true
		for channelID, packets := range channelPackets {
			if channelProcesed[channelID] != len(packets) {
				done = false
			}
		}
		if done {
			break
		}
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

// next block with res
func nextBlockWithRes(chain *ibctesting.TestChain, nextBlockTime time.Time) (abci.ResponseEndBlock, abci.ResponseBeginBlock) {
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

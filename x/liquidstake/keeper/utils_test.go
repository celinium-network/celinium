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

	"github.com/celinium-netwok/celinium/app/params"
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

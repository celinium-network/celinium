package keeper_test

import (
	"encoding/json"
	"testing"

	icaapp "celinium/app"
	"celinium/app/params"

	"celinium/x/inter-staking/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encCdc := icaapp.MakeEncodingConfig()

	app := icaapp.NewApp(log.NewNopLogger(), db, nil, true, nil, "", 0, encCdc, icaapp.EmptyAppOptions{})
	return app, icaapp.NewDefaultGenesisState(encCdc.Codec)
}

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// source chain
	sourceChain *ibctesting.TestChain
	// control chain
	controlChain *ibctesting.TestChain

	transferPath *ibctesting.Path

	interStakingPath *ibctesting.Path
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupSuite() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.sourceChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.controlChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	suite.transferPath = NewTransferPath(suite.sourceChain, suite.controlChain)
	suite.coordinator.Setup(suite.transferPath)

	suite.interStakingPath = NewICAPath(suite.sourceChain, suite.controlChain)
	suite.coordinator.SetupConnections(suite.interStakingPath)
}

func (suite *KeeperTestSuite) InterChainDelegate(
	ctlChain *ibctesting.TestChain, sourceChainID string, delegator sdk.AccAddress, coin sdk.Coin,
) []channeltypes.MsgRecvPacket {
	ctlChainApp := GetLocalApp(ctlChain)

	ctlChainCtx := ctlChain.GetContext()
	err := ctlChainApp.InterStakingKeeper.
		Delegate(
			ctlChainCtx,
			sourceChainID,
			coin,
			delegator.String(),
		)
	suite.NoError(err)

	res := ctlChainApp.EndBlock(abcitypes.RequestEndBlock{
		Height: ctlChainCtx.BlockHeight(),
	})

	sendPackets := parsePacketFromABCIEvents(res.Events)

	suite.coordinator.CommitBlock(suite.controlChain)

	var recvMsgs []channeltypes.MsgRecvPacket
	for _, p := range sendPackets {

		if p.DestinationPort == transfertypes.PortID {
			suite.transferPath.EndpointA.UpdateClient()
		} else {
			suite.interStakingPath.EndpointA.UpdateClient()
		}

		commitKey := ibchost.PacketCommitmentKey(p.SourcePort, p.SourceChannel, p.Sequence)
		proof, height := ctlChain.QueryProof(commitKey)
		backProofType := ibccommitmenttypes.MerkleProof{}
		backProofType.Unmarshal(proof)
		recvMsgs = append(recvMsgs, channeltypes.MsgRecvPacket{
			Packet:          p,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          delegator.String(),
		})
	}

	return recvMsgs
}

func (suite *KeeperTestSuite) TestDelegate() {
	var err error
	amount := math.NewIntFromUint64(1000)

	coin := sdk.NewCoin(params.DefaultBondDenom, amount)
	sourceChainUserAddress := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddress := suite.controlChain.SenderAccount.GetAddress()
	sourceChainApp := GetLocalApp(suite.sourceChain)
	ctlChainApp := GetLocalApp(suite.controlChain)

	mintCoin(suite.sourceChain, sourceChainUserAddress, coin)
	suite.CrossChainTransfer(coin, sourceChainUserAddress, controlChainUserAddress)

	traceCoin := sdk.NewCoin(firstIBCCoinDenom(suite.controlChain), coin.Amount)
	SetupInterStakingPath(suite.interStakingPath, traceCoin.Denom)

	sourceChainMetadata, _ := ctlChainApp.InterStakingKeeper.GetSourceChain(suite.controlChain.GetContext(), suite.sourceChain.ChainID)
	portID, _ := icatypes.NewControllerPortID(sourceChainMetadata.ICAControlAddr)
	hostAddr, _ := ctlChainApp.ICAControllerKeeper.GetInterchainAccountAddress(suite.controlChain.GetContext(), sourceChainMetadata.IbcConnectionId, portID)
	valAddress, err := sdk.ValAddressFromBech32(sourceChainMetadata.DelegateStrategy[0].ValidatorAddress)
	suite.Require().NoError(err)

	assembledRecvMsgs := suite.InterChainDelegate(suite.controlChain, suite.sourceChain.ChainID, controlChainUserAddress, traceCoin)

	var interstakingPathevents sdk.Events
	for i := 0; i < len(assembledRecvMsgs); i++ {
		ctx := suite.sourceChain.GetContext()
		_, err = sourceChainApp.IBCKeeper.RecvPacket(ctx, &assembledRecvMsgs[i])
		suite.Require().NoError(err)
		if assembledRecvMsgs[i].Packet.SourceChannel == suite.interStakingPath.EndpointA.ChannelID {
			interstakingPathevents = append(interstakingPathevents, ctx.EventManager().Events()...)
		}
	}

	_, found := sourceChainApp.StakingKeeper.GetDelegation(suite.sourceChain.GetContext(), sdk.MustAccAddressFromBech32(hostAddr), valAddress)
	suite.True(found)

	suite.sourceChain.NextBlock()
	suite.interStakingPath.EndpointB.UpdateClient()

	icaMsgRecvPacket := assembledRecvMsgs[1]

	key := ibchost.PacketAcknowledgementKey(icaMsgRecvPacket.Packet.GetDestPort(),
		icaMsgRecvPacket.Packet.GetDestChannel(),
		icaMsgRecvPacket.Packet.GetSequence())

	ackproof, ackheight := suite.sourceChain.QueryProof(key)
	ackProofType := ibccommitmenttypes.MerkleProof{}
	ackProofType.Unmarshal(ackproof)

	ackFromEvent, err := ibctesting.ParseAckFromEvents(interstakingPathevents)
	suite.Require().NoError(err)

	ackMsg := channeltypes.MsgAcknowledgement{
		Packet:          icaMsgRecvPacket.Packet,
		Acknowledgement: ackFromEvent,
		ProofAcked:      ackproof,
		ProofHeight:     ackheight,
		Signer:          controlChainUserAddress.String(),
	}

	_, err = ctlChainApp.IBCKeeper.Acknowledgement(suite.controlChain.GetContext(), &ackMsg)
	suite.Require().NoError(err)

	// check delegator's delegation
	delegationCoin := ctlChainApp.InterStakingKeeper.GetDelegation(suite.controlChain.GetContext(), controlChainUserAddress.String(), suite.sourceChain.ChainID)
	suite.Equal(delegationCoin.Amount, amount)
}

func mintCoin(chain *ibctesting.TestChain, to sdk.AccAddress, coin sdk.Coin) {
	sourceChainApp := GetLocalApp(chain)
	sourceChainApp.BankKeeper.
		MintCoins(
			chain.GetContext(),
			transfertypes.ModuleName,
			sdk.NewCoins(coin),
		)

	sourceChainApp.BankKeeper.
		SendCoinsFromModuleToAccount(
			chain.GetContext(),
			transfertypes.ModuleName,
			to,
			sdk.NewCoins(coin))
}

func firstIBCCoinDenom(chain *ibctesting.TestChain) string {
	traces := GetLocalApp(chain).TransferKeeper.GetAllDenomTraces(chain.GetContext())
	return traces[0].IBCDenom()
}

// CrossChainTransfer transfer coin from source chain to control chain
func (suite *KeeperTestSuite) CrossChainTransfer(coin sdk.Coin, from sdk.Address, to sdk.Address) {
	// sourceChainApp := GetLocalApp(suite.sourceChain)
	controlChainApp := GetLocalApp(suite.controlChain)

	// corss chain transfer from source chain to control chain.
	transferMsg := transfertypes.NewMsgTransfer(
		suite.transferPath.EndpointA.ChannelConfig.PortID,
		suite.transferPath.EndpointA.ChannelID,
		coin,
		from.String(),
		to.String(),
		suite.sourceChain.GetTimeoutHeight(),
		0,
		"",
	)
	// Begin send corss chain transfer message.
	// Because IBC cross-chain communication must obtain the proof of the source chain,
	// this proof can only be obtained in SendMsg???
	res, err := suite.sourceChain.SendMsgs(transferMsg)
	suite.Require().NoError(err)
	suite.transferPath.EndpointB.UpdateClient()

	// get corss chain transfer response from result of transaction.
	resp := transfertypes.MsgTransferResponse{}
	resp.Unmarshal(res.MsgResponses[0].Value)

	// Get cross chain transfer proof
	commitKey := ibchost.PacketCommitmentKey(suite.transferPath.EndpointA.ChannelConfig.PortID, suite.transferPath.EndpointA.ChannelID, resp.Sequence)
	proof, height := suite.sourceChain.QueryProof(commitKey)

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	msgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          packet,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          transferMsg.Sender,
	}

	// control chain recv cross chain transfer packet.
	_, err = controlChainApp.IBCKeeper.RecvPacket(suite.controlChain.GetContext(), &msgRecvPacket)
	suite.Require().NoError(err)
}

func NewICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	return path
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

// SetupInterStakingPath establishes interstaking relationship.
// ChainA as source chain.
// ChainB as Host chain.
func SetupInterStakingPath(path *ibctesting.Path, traceCoinDenom string) error {
	chainA := path.EndpointA.Chain
	chainB := path.EndpointB.Chain

	strategy := []types.DelegationStrategy{
		{
			Percentage:       100,
			ValidatorAddress: sdk.ValAddress(chainA.Vals.Validators[0].Address).String(),
		},
	}

	var icaCtlAddr string
	var err error
	channelSequence := chainB.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(chainB.GetContext())

	if icaCtlAddr, err = GetLocalApp(path.EndpointB.Chain).
		InterStakingKeeper.
		AddSourceChain(
			chainB.GetContext(),
			strategy,
			params.DefaultBondDenom,
			traceCoinDenom,
			chainA.ChainID,
			path.EndpointB.ConnectionID,
			"channel-0",
			"",
		); err != nil {
		return err
	}
	// commit state changes for proof verification
	chainB.NextBlock()

	// set portID/ChannelID for endpointB
	portID, err := icatypes.NewControllerPortID(icaCtlAddr)
	if err != nil {
		return err
	}
	path.EndpointB.ChannelConfig.PortID = portID
	path.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)

	// set channel version
	channel, _ := GetLocalApp(path.EndpointB.Chain).IBCKeeper.ChannelKeeper.GetChannel(chainB.GetContext(), portID, path.EndpointB.ChannelID)
	path.EndpointB.ChannelConfig.Version = channel.Version

	if err := path.EndpointA.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenAck(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenConfirm(); err != nil {
		return err
	}

	return nil
}

func GetLocalApp(chain *ibctesting.TestChain) *icaapp.App {
	app, ok := chain.App.(*icaapp.App)
	if !ok {
		panic("not ica app")
	}

	return app
}

func parsePacketFromABCIEvents(abciEvents []abcitypes.Event) []channeltypes.Packet {
	packets := make([]channeltypes.Packet, 0)
	for _, ev := range abciEvents {
		events := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}
		p, err := ibctesting.ParsePacketFromEvents(events)
		if err != nil {
			continue
		}
		packets = append(packets, p)
	}

	return packets
}

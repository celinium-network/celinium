package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	icaapp "celinium/app"
	"celinium/app/params"

	"celinium/x/inter-staking/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	icaapp.DefaultUnbondingTime = time.Minute * 5
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

func TestKeeperTest(t *testing.T) {
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

	/* --- Unelegate --- */
	ctlChainCtx := suite.controlChain.GetContext()

	err = ctlChainApp.InterStakingKeeper.
		UnDelegate(ctlChainCtx, suite.sourceChain.ChainID, traceCoin, controlChainUserAddress.String())
	suite.Require().NoError(err)
	undelegationEvents := ctlChainCtx.EventManager().Events()

	suite.coordinator.CommitBlock(suite.controlChain)
	suite.interStakingPath.EndpointA.UpdateClient()

	sendPacket, err := ibctesting.ParsePacketFromEvents(undelegationEvents)
	suite.Require().NoError(err)

	commitKey := ibchost.PacketCommitmentKey(sendPacket.SourcePort, sendPacket.SourceChannel, sendPacket.Sequence)
	proof, height := suite.controlChain.QueryProof(commitKey)
	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)
	undelegateRecvMsg := channeltypes.MsgRecvPacket{
		Packet:          sendPacket,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          controlChainUserAddress.String(),
	}

	souceChainCtx := suite.sourceChain.GetContext()
	_, err = sourceChainApp.IBCKeeper.RecvPacket(souceChainCtx, &undelegateRecvMsg)
	undelegateEvents := souceChainCtx.EventManager().Events()
	suite.Require().NoError(err)

	suite.sourceChain.NextBlock()
	suite.interStakingPath.EndpointB.UpdateClient()

	undelegateAckkey := ibchost.PacketAcknowledgementKey(undelegateRecvMsg.Packet.GetDestPort(),
		undelegateRecvMsg.Packet.GetDestChannel(),
		undelegateRecvMsg.Packet.GetSequence())

	undelegateAckproof, undelegateheight := suite.sourceChain.QueryProof(undelegateAckkey)

	undelegateAckFromEvent, err := ibctesting.ParseAckFromEvents(undelegateEvents)
	suite.Require().NoError(err)

	undelegateAckMsg := channeltypes.MsgAcknowledgement{
		Packet:          undelegateRecvMsg.Packet,
		Acknowledgement: undelegateAckFromEvent,
		ProofAcked:      undelegateAckproof,
		ProofHeight:     undelegateheight,
		Signer:          controlChainUserAddress.String(),
	}

	_, err = ctlChainApp.IBCKeeper.Acknowledgement(suite.controlChain.GetContext(), &undelegateAckMsg)
	suite.Require().NoError(err)

	hostAddress := sdk.MustAccAddressFromBech32(hostAddr)
	udbKey := stakingtypes.GetUBDKey(hostAddress, valAddress)

	suite.coordinator.IncrementTime()
	suite.coordinator.CommitBlock(suite.sourceChain)
	suite.interStakingPath.EndpointB.UpdateClient()

	ubdQueue, _ := sourceChainApp.StakingKeeper.GetUnbondingDelegation(suite.sourceChain.GetContext(), hostAddress, valAddress)
	ubdproof, height := QueryProofAtHeight(suite.sourceChain, stakingtypes.StoreKey, udbKey, suite.sourceChain.App.LastBlockHeight())

	xbackProofType := ibccommitmenttypes.MerkleProof{}
	xbackProofType.Unmarshal(ubdproof)

	err = ctlChainApp.InterStakingKeeper.SubmitSourceChainUnbondingDelegation(
		suite.controlChain.GetContext(),
		suite.interStakingPath.EndpointA.Chain.ChainID,
		suite.interStakingPath.EndpointA.ClientID,
		[][]byte{ubdproof},
		height,
		[]stakingtypes.UnbondingDelegation{ubdQueue},
	)
	suite.Require().NoError(err)

	for i := 0; i < 30; i++ {
		suite.interStakingPath.EndpointB.UpdateClient()
		suite.interStakingPath.EndpointA.UpdateClient()
		suite.transferPath.EndpointB.UpdateClient()
		suite.transferPath.EndpointA.UpdateClient()
	}

	suite.interStakingPath.EndpointB.UpdateClient()

	// _, found = sourceChainApp.StakingKeeper.GetUnbondingDelegation(suite.sourceChain.GetContext(), hostAddress, valAddress)
	ubdproof, height = QueryProofAtHeight(suite.sourceChain, stakingtypes.StoreKey, udbKey, suite.sourceChain.App.LastBlockHeight())

	xbackProofType = ibccommitmenttypes.MerkleProof{}
	xbackProofType.Unmarshal(ubdproof)

	ctlChainCtx = suite.controlChain.GetContext()

	err = ctlChainApp.InterStakingKeeper.SubmitSourceChainDVPairNotExist(
		ctlChainCtx,
		suite.interStakingPath.EndpointA.Chain.ChainID,
		suite.interStakingPath.EndpointA.ClientID,
		[][]byte{ubdproof},
		height,
		[]stakingtypes.DVPair{{
			DelegatorAddress: hostAddr,
			ValidatorAddress: valAddress.String(),
		}},
	)
	suite.Require().NoError(err)

	suite.coordinator.CommitBlock(suite.controlChain)
	suite.interStakingPath.EndpointA.UpdateClient()

	sp, err := ibctesting.ParsePacketFromEvents(ctlChainCtx.EventManager().Events())
	suite.Require().NoError(err)
	commitKey = ibchost.PacketCommitmentKey(sp.SourcePort, sp.SourceChannel, sp.Sequence)

	proof, height = suite.controlChain.QueryProof(commitKey)
	backProofType = ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)
	recvMsg := channeltypes.MsgRecvPacket{
		Packet:          sp,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          controlChainUserAddress.String(),
	}

	suite.interStakingPath.EndpointA.UpdateClient()
	balanceBefore := sourceChainApp.BankKeeper.GetBalance(suite.sourceChain.GetContext(), controlChainUserAddress, params.DefaultBondDenom)

	suite.coordinator.CommitBlock(suite.sourceChain)
	suite.transferPath.EndpointB.UpdateClient()

	sourceChainCtx := suite.sourceChain.GetContext()

	_, err = sourceChainApp.IBCKeeper.RecvPacket(sourceChainCtx, &recvMsg)
	suite.Require().NoError(err)

	p2, err := ibctesting.ParsePacketFromEvents(sourceChainCtx.EventManager().Events())
	suite.Require().NoError(err)

	suite.coordinator.CommitBlock(suite.sourceChain)
	// suite.transferPath.EndpointB.UpdateClient()
	// err = suite.transferPath.EndpointB.UpdateClient()
	// suite.Require().NoError(err)
	suite.transferPath.EndpointB.UpdateClient()

	commitKey2 := ibchost.PacketCommitmentKey(p2.SourcePort, p2.SourceChannel, p2.Sequence)

	proof, height = suite.sourceChain.QueryProof(commitKey2)
	backProofType = ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(proof)
	recvMsg2 := channeltypes.MsgRecvPacket{
		Packet:          p2,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          controlChainUserAddress.String(),
	}
	suite.transferPath.EndpointB.UpdateClient()

	_, err = ctlChainApp.IBCKeeper.RecvPacket(suite.controlChain.GetContext(), &recvMsg2)
	suite.Require().NoError(err)

	balanceAfter := ctlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddress, traceCoin.Denom)

	if balanceAfter.Amount.LT(balanceBefore.Amount) {
		panic("check balance failed after undelegate completely")
	}
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

	tmConfig := ibctesting.NewTendermintConfig()
	tmConfig.UnbondingPeriod = icaapp.DefaultUnbondingTime
	tmConfig.TrustingPeriod = icaapp.DefaultUnbondingTime - time.Second

	path.EndpointA.ClientConfig = tmConfig
	path.EndpointB.ClientConfig = tmConfig

	return path
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	tmConfig := ibctesting.NewTendermintConfig()
	tmConfig.UnbondingPeriod = icaapp.DefaultUnbondingTime
	tmConfig.TrustingPeriod = icaapp.DefaultUnbondingTime - time.Second

	path.EndpointA.ClientConfig = tmConfig
	path.EndpointB.ClientConfig = tmConfig

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

func QueryProofAtHeight(chain *ibctesting.TestChain, storePrefix string, key []byte, height int64) ([]byte, clienttypes.Height) {
	res := chain.App.Query(abcitypes.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storePrefix),
		Height: height - 1,
		Data:   key,
		Prove:  true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.T, err)

	proof, err := chain.App.AppCodec().Marshal(&merkleProof)
	require.NoError(chain.T, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height)+1)
}

package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"

	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	icaapp "celinium/app"
	"celinium/app/params"
	"celinium/x/inter-staking/types"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"

	// TestPortID defines a reusable port identifier for testing purposes
	TestPortID, _ = icatypes.NewControllerPortID(TestOwnerAddress)
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

func (suite *KeeperTestSuite) TestDelegate() {
	amount := math.NewIntFromUint64(1000)

	coin := sdk.NewCoin(params.DefaultBondDenom, amount)
	sourceChainUserAddress := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddress := suite.controlChain.SenderAccount.GetAddress()

	mintCoin(suite.sourceChain, sourceChainUserAddress, coin)

	sourceChainApp := GetLocalApp(suite.sourceChain)

	resp, err := sourceChainApp.BankKeeper.Balance(suite.sourceChain.GetContext(), banktypes.NewQueryBalanceRequest(sourceChainUserAddress, coin.Denom))
	suite.Require().NoError(err)
	suite.Equal(*resp.Balance, coin)

	suite.CrossChainTransferForward(coin, sourceChainUserAddress, controlChainUserAddress)

	resp, err = sourceChainApp.BankKeeper.Balance(suite.sourceChain.GetContext(), banktypes.NewQueryBalanceRequest(sourceChainUserAddress, coin.Denom))
	suite.Require().NoError(err)
	suite.Equal(resp.Balance.Amount, math.ZeroInt())

	traceCoin := sdk.NewCoin(firstIBCCoinDenom(suite.controlChain), coin.Amount)
	SetupInterStakingPath(suite.interStakingPath, traceCoin.Denom)

	// suite.CrossChainTransferBack(traceCoin, controlChainUserAddress, sourceChainUserAddress)
	controlChainApp := GetLocalApp(suite.controlChain)
	controlChainApp.InterStakingKeeper.Delegate(suite.controlChain.GetContext(), suite.sourceChain.ChainID, traceCoin, controlChainUserAddress.String())
	suite.transferPath.EndpointA.UpdateClient()
	suite.transferPath.EndpointA.UpdateClient()

	/*** relay transfer msg **/
	backCommitKey := ibchost.PacketCommitmentKey(suite.transferPath.EndpointB.ChannelConfig.PortID, suite.transferPath.EndpointB.ChannelID, 1)
	backproof, backheight := suite.controlChain.QueryProof(backCommitKey)
	fmt.Println(backproof, backheight)
	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(backproof)

	sourceChainMetadata, _ := controlChainApp.InterStakingKeeper.GetSourceChain(suite.controlChain.GetContext(), suite.sourceChain.ChainID)
	portID, _ := icatypes.NewControllerPortID(sourceChainMetadata.ICAControlAddr)
	hostAddr, _ := controlChainApp.ICAControllerKeeper.GetInterchainAccountAddress(suite.controlChain.GetContext(), sourceChainMetadata.IbcConnectionId, portID)

	fullDenomPath, err := controlChainApp.TransferKeeper.DenomPathFromHash(suite.controlChain.GetContext(), traceCoin.Denom)
	suite.Require().NoError(err)

	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    fullDenomPath,
		Amount:   traceCoin.Amount.String(),
		Sender:   sourceChainMetadata.ICAControlAddr,
		Receiver: hostAddr,
		Memo:     "",
	}

	backChannelPacket := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         suite.transferPath.EndpointB.ChannelConfig.PortID,
		SourceChannel:      suite.transferPath.EndpointB.ChannelID,
		DestinationPort:    suite.transferPath.EndpointA.ChannelConfig.PortID,
		DestinationChannel: suite.transferPath.EndpointA.ChannelID,
		Data:               packetData.GetBytes(),
		TimeoutHeight:      clienttypes.NewHeight(0, 10000),
		TimeoutTimestamp:   0,
	}

	backMsgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          backChannelPacket,
		ProofCommitment: backproof,
		ProofHeight:     backheight,
		Signer:          controlChainUserAddress.String(),
	}

	beforeResp, _ := sourceChainApp.BankKeeper.Balance(suite.sourceChain.GetContext(), banktypes.NewQueryBalanceRequest(sdk.MustAccAddressFromBech32(hostAddr), coin.Denom))
	_, err = sourceChainApp.IBCKeeper.RecvPacket(suite.sourceChain.GetContext(), &backMsgRecvPacket)
	suite.Require().NoError(err)

	// check balance
	afrerResp, err := sourceChainApp.BankKeeper.Balance(suite.sourceChain.GetContext(), banktypes.NewQueryBalanceRequest(sdk.MustAccAddressFromBech32(hostAddr), coin.Denom))
	suite.Require().NoError(err)
	suite.Equal(afrerResp.Balance.Amount.Sub(beforeResp.Balance.Amount), packetData.Amount)

	// relay staking tx
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

// CrossChainTransferForward transfer coin from source chain to control chain
func (suite *KeeperTestSuite) CrossChainTransferForward(coin sdk.Coin, from sdk.Address, to sdk.Address) {
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

	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    transferMsg.Token.Denom,
		Amount:   transferMsg.Token.Amount.String(),
		Sender:   transferMsg.Sender,
		Receiver: transferMsg.Receiver,
		Memo:     transferMsg.Memo,
	}

	channelPacket := channeltypes.Packet{
		Sequence:           resp.Sequence,
		SourcePort:         transferMsg.SourcePort,
		SourceChannel:      transferMsg.SourceChannel,
		DestinationPort:    suite.transferPath.EndpointB.ChannelConfig.PortID,
		DestinationChannel: suite.transferPath.EndpointB.ChannelID,
		Data:               packetData.GetBytes(),
		TimeoutHeight:      transferMsg.TimeoutHeight,
		TimeoutTimestamp:   transferMsg.TimeoutTimestamp,
	}

	msgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          channelPacket,
		ProofCommitment: proof,
		ProofHeight:     height,
		Signer:          transferMsg.Sender,
	}

	// control chain recv cross chain transfer packet.
	_, err = controlChainApp.IBCKeeper.RecvPacket(suite.controlChain.GetContext(), &msgRecvPacket)
	suite.Require().NoError(err)
}

// CrossChainTransferBack transfer coin from control to source chain
func (suite *KeeperTestSuite) CrossChainTransferBack(coin sdk.Coin, from sdk.Address, to sdk.Address) {
	sourceChainApp := GetLocalApp(suite.sourceChain)
	controlChainApp := GetLocalApp(suite.controlChain)

	transferBackMsg := transfertypes.NewMsgTransfer(
		suite.transferPath.EndpointB.ChannelConfig.PortID,
		suite.transferPath.EndpointB.ChannelID,
		coin,
		from.String(),
		to.String(),
		suite.controlChain.GetTimeoutHeight(),
		0,
		"",
	)
	transferBackRes, err := suite.controlChain.SendMsgs(transferBackMsg)
	suite.Require().NoError(err)
	suite.transferPath.EndpointA.UpdateClient()

	backResp := transfertypes.MsgTransferResponse{}
	backResp.Unmarshal(transferBackRes.MsgResponses[0].Value)
	backCommitKey := ibchost.PacketCommitmentKey(suite.transferPath.EndpointB.ChannelConfig.PortID, suite.transferPath.EndpointB.ChannelID, backResp.Sequence)
	backproof, backheight := suite.controlChain.QueryProof(backCommitKey)
	backProofType := ibccommitmenttypes.MerkleProof{}
	backProofType.Unmarshal(backproof)

	fullDenomPath, err := controlChainApp.TransferKeeper.DenomPathFromHash(suite.controlChain.GetContext(), transferBackMsg.Token.Denom)
	suite.Require().NoError(err)
	backpacketData := transfertypes.FungibleTokenPacketData{
		Denom:    fullDenomPath,
		Amount:   transferBackMsg.Token.Amount.String(),
		Sender:   transferBackMsg.Sender,
		Receiver: transferBackMsg.Receiver,
		Memo:     transferBackMsg.Memo,
	}

	backChannelPacket := channeltypes.Packet{
		Sequence:           backResp.Sequence,
		SourcePort:         transferBackMsg.SourcePort,
		SourceChannel:      transferBackMsg.SourceChannel,
		DestinationPort:    suite.transferPath.EndpointA.ChannelConfig.PortID,
		DestinationChannel: suite.transferPath.EndpointA.ChannelID,
		Data:               backpacketData.GetBytes(),
		TimeoutHeight:      transferBackMsg.TimeoutHeight,
		TimeoutTimestamp:   transferBackMsg.TimeoutTimestamp,
	}

	backMsgRecvPacket := channeltypes.MsgRecvPacket{
		Packet:          backChannelPacket,
		ProofCommitment: backproof,
		ProofHeight:     backheight,
		Signer:          transferBackMsg.Sender,
	}

	_, err = sourceChainApp.IBCKeeper.RecvPacket(suite.sourceChain.GetContext(), &backMsgRecvPacket)
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
			ValidatorAddress: chainA.Vals.Validators[0].Address.String(),
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
			traceCoinDenom,
			chainA.ChainID,
			path.EndpointB.ConnectionID,
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

func assembleChannelVersion(ctlConnID, hostConnID string) string {
	return string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ctlConnID,
		HostConnectionId:       hostConnID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
}

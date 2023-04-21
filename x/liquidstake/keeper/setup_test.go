package keeper_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	celiniumapp "github.com/celinium-network/celinium/app"
	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	"github.com/celinium-network/celinium/x/liquidstake/keeper"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	transferPath *ibctesting.Path
	icaPath      *ibctesting.Path

	sourceChain  *ibctesting.TestChain
	controlChain *ibctesting.TestChain

	testCoin sdk.Coin

	queryClient types.QueryClient
}

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encCdc := celiniumapp.MakeEncodingConfig()

	app := celiniumapp.NewApp(log.NewNopLogger(), db, nil, true, nil, "", 0, encCdc, celiniumapp.EmptyAppOptions{})

	// replace production epoch module genesis by default epoch module genesis
	genesisState := celiniumapp.NewDefaultGenesisState(encCdc.Codec)
	genesisState[epochtypes.ModuleName] = encCdc.Codec.MustMarshalJSON(epochtypes.DefaultGenesisState())

	return app, genesisState
}

func TestKeeperTestSuite(t *testing.T) {
	s := new(KeeperTestSuite)
	suite.Run(t, s)
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.sourceChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.controlChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	suite.transferPath = newTransferPath(suite.sourceChain, suite.controlChain)
	suite.coordinator.Setup(suite.transferPath)

	suite.icaPath = newICAPath(suite.sourceChain, suite.controlChain)
	suite.icaPath = copyConnectionAndClientToPath(suite.icaPath, suite.transferPath)

	// create test coin, mint in sourcechain, then transfer to control chain by ibc
	srcChainParams := suite.generateSourceChainParams()

	amount := uint64(rand.Int()%100000 + 100000) //nolint:gosec
	testCoin := sdk.NewCoin(srcChainParams.NativeDenom, sdk.NewIntFromUint64(amount))

	srcChainUserAccAddr := suite.sourceChain.SenderAccount.GetAddress()
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()

	mintCoin(suite.sourceChain, srcChainUserAccAddr, testCoin)
	suite.IBCTransfer(srcChainUserAccAddr.String(), ctlChainUserAccAddr.String(), testCoin, suite.transferPath, true)

	suite.testCoin = testCoin

	// create liquidstake query client
	queryHelper := baseapp.NewQueryServerTestHelper(suite.controlChain.GetContext(),
		getCeliniumApp(suite.controlChain).InterfaceRegistry())
	querier := keeper.Querier{Keeper: getCeliniumApp(suite.controlChain).LiquidStakeKeeper}
	types.RegisterQueryServer(queryHelper, querier)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func newTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	tmConfig := ibctesting.NewTendermintConfig()
	tmConfig.UnbondingPeriod = celiniumapp.DefaultUnbondingTime
	tmConfig.TrustingPeriod = celiniumapp.DefaultUnbondingTime - time.Minute*5
	tmConfig.MaxClockDrift = celiniumapp.DefaultUnbondingTime - time.Minute*5

	path.EndpointA.ClientConfig = tmConfig
	path.EndpointB.ClientConfig = tmConfig

	return path
}

func newICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	tmConfig := ibctesting.NewTendermintConfig()
	tmConfig.UnbondingPeriod = celiniumapp.DefaultUnbondingTime
	tmConfig.TrustingPeriod = celiniumapp.DefaultUnbondingTime - time.Minute*5
	tmConfig.MaxClockDrift = celiniumapp.DefaultUnbondingTime - time.Minute*5

	path.EndpointA.ClientConfig = tmConfig
	path.EndpointB.ClientConfig = tmConfig

	return path
}

func copyConnectionAndClientToPath(path *ibctesting.Path, pathToCopy *ibctesting.Path) *ibctesting.Path {
	path.EndpointA.ClientID = pathToCopy.EndpointA.ClientID
	path.EndpointB.ClientID = pathToCopy.EndpointB.ClientID
	path.EndpointA.ConnectionID = pathToCopy.EndpointA.ConnectionID
	path.EndpointB.ConnectionID = pathToCopy.EndpointB.ConnectionID
	path.EndpointA.ClientConfig = pathToCopy.EndpointA.ClientConfig
	path.EndpointB.ClientConfig = pathToCopy.EndpointB.ClientConfig
	path.EndpointA.ConnectionConfig = pathToCopy.EndpointA.ConnectionConfig
	path.EndpointB.ConnectionConfig = pathToCopy.EndpointB.ConnectionConfig
	return path
}

func getCeliniumApp(chain *ibctesting.TestChain) *celiniumapp.App {
	app, ok := chain.App.(*celiniumapp.App)
	if !ok {
		panic("not celinium app")
	}

	return app
}

func (suite *KeeperTestSuite) calcuateIBCDenom(portID, channelID string, denom string) string {
	sourcePrefix := transfertypes.GetDenomPrefix(portID, channelID)
	denomTrace := transfertypes.ParseDenomTrace(sourcePrefix + denom)

	return denomTrace.IBCDenom()
}

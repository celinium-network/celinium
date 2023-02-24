package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"celinium/app"
	"celinium/x/swap/keeper"
	"celinium/x/swap/types"
)

var (
	secondaryDenom  = "uion"
	secondaryAmount = sdk.NewInt(100000000)
)

type IntegrationTestSuite struct {
	suite.Suite

	App *app.App
	Ctx sdk.Context

	queryClient types.QueryClient
	msgServer   types.MsgServer

	TestAccs []sdk.AccAddress
}

func (suite *IntegrationTestSuite) SetupTest() {
	createdApp := app.Setup(suite.T(), false)
	ctx := createdApp.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	createdApp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	createdApp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, createdApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, createdApp.SwapKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = createdApp
	suite.Ctx = ctx
	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(createdApp.SwapKeeper)
	suite.TestAccs = app.AddTestAddrs(createdApp, ctx, 3, sdkmath.NewInt(1000000000))

	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(secondaryDenom, secondaryAmount))
	for _, acc := range suite.TestAccs {
		banktestutil.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, fundAccsAmount)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

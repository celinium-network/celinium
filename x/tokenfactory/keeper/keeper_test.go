package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"celinium/app"
	"celinium/x/tokenfactory/keeper"
	"celinium/x/tokenfactory/types"
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
	// defaultDenom is on the suite, as it depends on the creator test address.
	defaultDenom string
}

func (suite *IntegrationTestSuite) SetupTest() {
	createdApp := app.Setup(suite.T(), false)
	ctx := createdApp.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	createdApp.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	createdApp.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, createdApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, createdApp.TokenFactoryKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = createdApp
	suite.Ctx = ctx
	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(createdApp.TokenFactoryKeeper)
	suite.TestAccs = app.AddTestAddrs(createdApp, ctx, 3, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100))

	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(secondaryDenom, secondaryAmount))
	for _, acc := range suite.TestAccs {
		banktestutil.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, fundAccsAmount)
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) TestCreateModuleAccount() {
	app := suite.App

	// remove module account
	tokenfactoryModuleAccount := app.AccountKeeper.GetAccount(suite.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	app.AccountKeeper.RemoveAccount(suite.Ctx, tokenfactoryModuleAccount)

	// ensure module account was removed
	suite.Ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(suite.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	suite.Require().Nil(tokenfactoryModuleAccount)

	// create module account
	app.TokenFactoryKeeper.CreateModuleAccount(suite.Ctx)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = app.AccountKeeper.GetAccount(suite.Ctx, app.AccountKeeper.GetModuleAddress(types.ModuleName))
	suite.Require().NotNil(tokenfactoryModuleAccount)
}

func (suite *IntegrationTestSuite) CreateDefaultDenom() {
	res, _ := suite.msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.defaultDenom = res.GetNewTokenDenom()
}

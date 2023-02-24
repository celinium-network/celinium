package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"celinium/app"
	"celinium/app/params"
	"celinium/x/tokenfactory/client/cli"
	"celinium/x/tokenfactory/types"
)

const (
	defaultCreatedDenom = "defaultCreatedDenom"
	defaultUnit         = 1000000
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	encCfg := app.MakeEncodingConfig()

	cfg.Codec = encCfg.Codec
	cfg.TxConfig = encCfg.TxConfig
	cfg.LegacyAmino = encCfg.Amino
	cfg.InterfaceRegistry = encCfg.InterfaceRegistry
	cfg.AppConstructor = NewAppConstructor(encCfg)
	cfg.BondDenom = params.DefaultBondDenom
	cfg.MinGasPrices = fmt.Sprintf("0.000006%s", params.DefaultBondDenom)
	cfg.GenesisState = app.ModuleBasics.DefaultGenesis(encCfg.Codec)

	sg := stakingtypes.DefaultGenesisState()
	sg.Params.BondDenom = params.DefaultBondDenom
	cfg.GenesisState[stakingtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(sg)

	cfg.NumValidators = 2

	suite.Run(t, NewIntegrationTestSuite(cfg))
}

func NewAppConstructor(encodingCfg params.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		db := dbm.NewMemDB()
		app := app.NewApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 5, encodingCfg, app.EmptyAppOptions{})
		return app
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	fmt.Println("setup suite")
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestCreateDenomTx() string {
	s.network.WaitForNextBlock()

	val := s.network.Validators[0]

	subDenom := "abc"
	args := []string{
		val.Address.String(),
		subDenom,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCreateDenomTxCmd(), args)
	s.Require().NoError(err)

	return fmt.Sprintf("factory/%s/%s", val.Address.String(), subDenom)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestMintTx() {
	s.network.WaitForNextBlock()

	var err error
	val := s.network.Validators[0]
	args := []string{
		val.Address.String(),
		defaultCreatedDenom,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCreateDenomTxCmd(), args)
	s.Require().NoError(err)

	mintArgs := []string{
		val.Address.String(),
		fmt.Sprintf("%dfactory/%s/%s", 100*defaultUnit, val.Address.String(), "defaultCreatedDenom"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewMintTxCmd(), mintArgs)
	s.Require().NoError(err)

}

func (s *IntegrationTestSuite) TestBurnTx() {
	s.TestMintTx()

	s.network.WaitForNextBlock()

	var err error
	val := s.network.Validators[0]
	mintArgs := []string{
		val.Address.String(),
		fmt.Sprintf("%dfactory/%s/%s", 50*defaultUnit, val.Address.String(), "defaultCreatedDenom"),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewBurnTxCmd(), mintArgs)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestChangeAdmin() {
	createdDenom := s.TestCreateDenomTx()

	s.network.WaitForNextBlock()

	testAddr := sdk.MustAccAddressFromBech32("demo1c2lmjndackh52hpgfjzpswquqr4wcvlgprhrdd")

	var err error
	val := s.network.Validators[0]
	changeAdmin := []string{
		val.Address.String(),
		createdDenom,
		testAddr.String(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	_, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewChangeAdminCmd(), changeAdmin)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestCmdQueryDenomAuthorityMetadata() {
	val := s.network.Validators[0]
	denom := s.TestCreateDenomTx()

	query := []string{
		denom,
	}
	bz, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCmdQueryDenomAuthorityMetadata(), query)
	s.Require().NoError(err)

	response := types.QueryDenomAuthorityMetadataResponse{}
	s.cfg.Codec.MustUnmarshalJSON(bz.Bytes(), &response)
}

func (s *IntegrationTestSuite) TestCmdQueryDenomsFromCreator() {
	val := s.network.Validators[0]

	query := []string{
		val.Address.String(),
	}

	bz, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.NewCmdQueryDenomsFromCreator(), query)
	s.Require().NoError(err)

	response := types.QueryDenomsFromCreatorResponse{}
	s.cfg.Codec.MustUnmarshalJSON(bz.Bytes(), &response)
}

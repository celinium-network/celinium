package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/app"
	"github.com/celinium-network/celinium/testutil"

	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App
	// queryClient types.QueryClient
	consAddress sdk.ConsAddress
}

var _ *KeeperTestSuite

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest()
}

func (suite *KeeperTestSuite) DoSetupTest() {
	checkTx := false

	// init app
	suite.app = app.Setup(suite.T(), checkTx)

	// setup context
	header := testutil.NewHeader(
		1, time.Now().UTC(), "test", suite.consAddress, nil, nil,
	)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, header)

	epochTemplate := epochtypes.EpochInfo{
		StartTime:               suite.ctx.BlockTime(),
		CurrentEpoch:            0,
		CurrentEpochStartTime:   suite.ctx.BlockTime(),
		EpochCountingStarted:    false,
		CurrentEpochStartHeight: suite.ctx.BlockHeight(),
	}

	refreshAgentDelegationEpoch := epochTemplate
	refreshAgentDelegationEpoch.Identifier = types.RefreshAgentDelegationEpochID
	refreshAgentDelegationEpoch.Duration = time.Hour

	collectAgentStakingRewardEpoch := epochTemplate
	collectAgentStakingRewardEpoch.Identifier = types.CollectAgentStakingRewardEpochID
	collectAgentStakingRewardEpoch.Duration = time.Hour * 2

	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, refreshAgentDelegationEpoch)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, collectAgentStakingRewardEpoch)
}

func TestKeeperTestSuite(t *testing.T) {
	s := new(KeeperTestSuite)
	suite.Run(t, s)
}

func (suite *KeeperTestSuite) mintCoin(coin sdk.Coin, recipient sdk.AccAddress) {
	suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.Coins{coin})
	err := suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, recipient, sdk.Coins{coin})
	suite.Require().NoError(err)
}

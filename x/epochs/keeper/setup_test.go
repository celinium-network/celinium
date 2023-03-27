package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"celinium/app"
	"celinium/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	queryClient types.QueryClient
	consAddress sdk.ConsAddress
}

var s *KeeperTestSuite

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest()
}

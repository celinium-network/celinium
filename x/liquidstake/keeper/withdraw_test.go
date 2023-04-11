package keeper_test

import (
	"fmt"
	"time"

	"github.com/celinium-netwok/celinium/app"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestWithdrawCompleteUnbond() {
	sourceChainParams := suite.mockSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(sourceChainParams, delegationEpochInfo)

	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	testCoin := sdk.NewCoin(sourceChainParams.NativeDenom, sdk.NewIntFromUint64(100000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, testCoin)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), testCoin, suite.transferPath, true)

	controlChainApp := getCeliniumApp(suite.controlChain)

	err := controlChainApp.LiquidStakeKeeper.Delegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)
	suite.NoError(err)

	suite.processDelegation(delegationEpochInfo)

	// user has already delegate, then undelegate
	unbondingEpochInfo := suite.unbondingEpoch()
	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	controlChainApp.LiquidStakeKeeper.Undelegate(suite.controlChain.GetContext(), sourceChainParams.ChainID, testCoin.Amount, controlChainUserAddr)

	// process at next unbond epoch begin
	coordTime := suite.advanceToNextEpoch(unbondingEpochInfo)

	nextBlockTime := coordTime
	_, nextBlockBeginRes := nextBlockWithRes(suite.controlChain, nextBlockTime)
	nextBlockWithRes(suite.sourceChain, nextBlockTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()

	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, controlChainUserAddr.String())

	epochUnbonding, found := controlChainApp.LiquidStakeKeeper.GetEpochUnboundings(suite.controlChain.GetContext(), 2)
	suite.True(found)

	var unbondCompleteTime time.Time
	for _, unbonding := range epochUnbonding.Unbondings {
		if unbonding.ChainID != sourceChainParams.ChainID {
			continue
		}
		suite.Equal(unbonding.Status, uint32(types.UnbondingWaitting))
		unbondCompleteTime = time.Unix(0, int64(unbonding.UnbondTIme))
	}

	// make the light client not expired.
	stepDuration := (time.Hour * 24)
	step := app.DefaultUnbondingTime / (time.Hour * 24)
	for i := 0; i < int(step-1); i++ {
		suite.controlChain.Coordinator.CurrentTime = suite.controlChain.Coordinator.CurrentTime.Add(stepDuration)

		fmt.Println(suite.sourceChain.Coordinator.CurrentTime.Format(time.RFC3339))

		nextBlockWithRes(suite.controlChain, suite.controlChain.Coordinator.CurrentTime)
		nextBlockWithRes(suite.sourceChain, suite.sourceChain.Coordinator.CurrentTime)

		suite.transferPath.EndpointA.UpdateClient()
		suite.transferPath.EndpointB.UpdateClient()
	}

	unbondCompleteTime = unbondCompleteTime.Add(time.Minute * 20)

	nextBlockWithRes(suite.sourceChain, unbondCompleteTime)
	fmt.Println(unbondCompleteTime.Format(time.RFC3339))
	fmt.Println(suite.sourceChain.Coordinator.CurrentTime.Format(time.RFC3339))
	_, nextBlockBeginRes = nextBlockWithRes(suite.controlChain, unbondCompleteTime)

	suite.controlChain.NextBlock()
	suite.transferPath.EndpointA.UpdateClient()
	suite.controlChain.Coordinator.CurrentTime = unbondCompleteTime

	suite.relayIBCPacketFromCtlToSrc(nextBlockBeginRes.Events, controlChainUserAddr.String())
}

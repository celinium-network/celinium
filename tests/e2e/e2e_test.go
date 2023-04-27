package e2e

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestLiquidStake() {
	regparams := s.mockRegisterSourceChain()
	delegateAmt := math.NewIntFromUint64(10000000000)

	s.IBCTokenTransfer(delegateAmt)

	srcChain, err := s.LiquidStakeAddSourceChain(regparams)
	s.Require().NoError(err, fmt.Sprintf("liquidstake register source chain failed %v", err))

	s.LiquistakeDelegate(srcChain, delegateAmt)

	s.CheckChainDelegate(srcChain)

	rewardAmount := s.LiquidstakeReinvest(srcChain)

	s.CheckChainReinvest(srcChain, delegateAmt, rewardAmount)

	undelegateEpoch := s.LiquistakeUndelegate(srcChain, delegateAmt, rewardAmount)

	s.LiquidstakeClaim(sdk.NewCoin(srcChain.IbcDenom, delegateAmt.Add(rewardAmount)), srcChain.ChainID, undelegateEpoch)
}

package e2e

import (
	"fmt"

	"cosmossdk.io/math"
)

func (s *IntegrationTestSuite) TestLiquidStake() {
	regparams := s.mockRegisterSourceChain()
	delegateAmt := math.NewIntFromUint64(10000000000)

	s.IBCTokenTransfer(delegateAmt)

	srcChain, err := s.LiquidStakeAddSourceChain(regparams)
	s.Require().NoError(err, fmt.Sprintf("liquidstake register source chain failed %v", err))

	s.LiquistakeDelegate(srcChain, delegateAmt)

	s.CheckChainDelegate(srcChain)

	s.CheckChainReinvest(srcChain)
}

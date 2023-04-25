package e2e

import (
	"fmt"
)

func (s *IntegrationTestSuite) TestLiquidStake() {
	regparams := s.mockRegisterSourceChain()

	_, err := s.LiquidStakeAddSourceChain(regparams)
	s.Require().NoError(err, fmt.Sprintf("liquidstake register source chain failed %v", err))
}

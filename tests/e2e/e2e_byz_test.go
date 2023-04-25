package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// validatorMissBlocks
func (s *IntegrationTestSuite) testValidatorMissBlocks() error { //nolint:unused // this is called during e2e tests
	c := s.srcChain
	badValIndex := 1
	resource := s.valResources[c.ID][badValIndex]
	s.dkrPool.Client.StopContainer(resource.Container.ID, 5)

	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))

	accAddr, err := c.validators[badValIndex].keyRecord.GetAddress()
	s.NoError(err)

	valAddr := sdk.ValAddress(accAddr)
	valBeforeSlash, err := queryValidator(s.ctlChain.encfg.Codec, chainBAPIEndpoint, valAddr.String())
	s.NoError(err)

	time.Sleep(time.Minute * 5)

	valAfterSlash, err := queryValidator(s.ctlChain.encfg.Codec, chainBAPIEndpoint, valAddr.String())
	s.NoError(err)

	s.Require().True(valAfterSlash.Tokens.LT(valBeforeSlash.Tokens), "checkslash amount failed")

	return nil
}

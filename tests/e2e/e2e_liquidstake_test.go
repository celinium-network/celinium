package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	liquistakecli "github.com/celinium-network/celinium/x/liquidstake/client/cli"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (s *IntegrationTestSuite) TestLiquidStakeAddSourceChain() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var selectVals liquistakecli.CliValidators

	for _, v := range s.srcChain.validators {
		accAddr, _ := v.keyRecord.GetAddress()
		valAddr := sdk.ValAddress(accAddr)

		selectVals.Vals = append(selectVals.Vals, types.Validator{
			Address:          valAddr.String(),
			DelegationAmount: math.ZeroInt(),
			Weight:           1000000,
		})
	}

	selectValsBz, err := json.Marshal(selectVals)
	s.NoError(err)

	senderAccAddr, err := s.ctlChain.validators[0].keyRecord.GetAddress()
	s.NoError(err)

	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)
	liuidstakeCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"register-source-chain",
		s.srcChain.ID,
		"connection-0",
		"channel-0",
		"celivaloper",
		string(selectValsBz),
		"CELI",
		"vpCELI",
		fmt.Sprintf("--from=%s", senderAccAddr.String()),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, gas*10),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.ctlChain.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))

	resp, err := queryLiquidstakeSourceChain(s.ctlChain.encfg.Codec, chainBAPIEndpoint, s.srcChain.ID)
	s.NoError(err)
	// TODO more check
	s.Equal(resp.SourceChain.ChainID, s.srcChain.ID)
}



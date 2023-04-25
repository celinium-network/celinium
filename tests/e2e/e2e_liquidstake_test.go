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

type sourceChainParams struct {
	ChainID         string
	ConnectionID    string
	ChannelID       string
	ValPrefix       string
	CliVals         liquistakecli.CliValidators
	NativeDeonm     string
	DerivativeDenom string
	registor        string
}

func (s *IntegrationTestSuite) mockRegisterSourceChain() *sourceChainParams {
	regparams := sourceChainParams{
		ChainID:         s.srcChain.ID,
		ConnectionID:    "connection-0",
		ChannelID:       "channel-0",
		ValPrefix:       "celivaloper",
		NativeDeonm:     "CELI",
		DerivativeDenom: "vpCELI",
	}

	registor, err := s.ctlChain.validators[0].keyRecord.GetAddress()
	s.NoError(err)

	regparams.registor = registor.String()

	for _, v := range s.srcChain.validators {
		accAddr, _ := v.keyRecord.GetAddress()
		valAddr := sdk.ValAddress(accAddr)

		regparams.CliVals.Vals = append(regparams.CliVals.Vals, types.Validator{
			Address:          valAddr.String(),
			DelegationAmount: math.ZeroInt(),
			Weight:           1000000,
		})
	}

	return &regparams
}

func (s *IntegrationTestSuite) LiquidStakeAddSourceChain(regparams *sourceChainParams) (*types.SourceChain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	selectValsBz, err := json.Marshal(regparams.CliVals)
	if err != nil {
		return nil, err
	}

	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)
	liuidstakeCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"register-source-chain",
		regparams.ChainID,
		regparams.ConnectionID,
		regparams.ChannelID,
		regparams.ValPrefix,
		string(selectValsBz),
		regparams.NativeDeonm,
		regparams.DerivativeDenom,
		fmt.Sprintf("--from=%s", regparams.registor),
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
	if err != nil {
		return nil, err
	}

	// TODO check registe result ?
	return &resp.SourceChain, nil
}

func (s *IntegrationTestSuite) Delegate(chainID string) {
}

//nolint:unused
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ory/dockertest/v3/docker"
)

const (
	flagFrom            = "from"
	flagHome            = "home"
	flagFees            = "fees"
	flagGas             = "gas"
	flagOutput          = "output"
	flagChainID         = "chain-id"
	flagSpendLimit      = "spend-limit"
	flagGasAdjustment   = "gas-adjustment"
	flagFeeAccount      = "fee-account"
	flagBroadcastMode   = "broadcast-mode"
	flagKeyringBackend  = "keyring-backend"
	flagAllowedMessages = "allowed-messages"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(c *chain, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagGasAdjustment:  "1.5",
		flagChainID:        c.ID,
		flagHome:           celiniumHomePath,
		flagFees:           sdk.NewCoin(c.Denom, standardFeeAmount).String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) execEncode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("%s - Executing gaiad encoding with %v", c.ID, txPath)
	command := []string{
		c.ChainNodeBinary,
		txCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	s.executeCeliniumTxCommand(ctx, c, command, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.Logf("successfully encode with %v", txPath)
	return encoded
}

func (s *IntegrationTestSuite) execDecode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("%s - Executing gaiad decoding with %v", c.ID, txPath)
	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	s.executeCeliniumTxCommand(ctx, c, chainCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.Logf("successfully decode %v", txPath)
	return decoded
}

func (s *IntegrationTestSuite) execVestingTx(
	c *chain,
	method string,
	args []string,
	opt ...flagOption,
) {
	opts := applyOptions(c, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("%s - Executing gaiad %s with %v", c.ID, method, args)
	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		vestingtypes.ModuleName,
		method,
		"-y",
	}
	chainCommand = append(chainCommand, args...)

	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, 0, s.defaultExecValidation(c, 0))
	s.Logf("successfully %s with %v", method, args)
}

func (s *IntegrationTestSuite) execCreatePeriodicVestingAccount(
	c *chain,
	address,
	jsonPath string,
	opt ...flagOption,
) {
	s.Logf("Executing gaiad create periodic vesting account %s", c.ID)
	s.execVestingTx(c, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	s.Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}

func (s *IntegrationTestSuite) execUnjail(
	c *chain,
	opt ...flagOption,
) {
	opts := applyOptions(c, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Executing gaiad slashing unjail %s with options: %v", c.ID, opt)
	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		slashingtypes.ModuleName,
		"unjail",
		"-y",
	}

	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, 0, s.defaultExecValidation(c, 0))
	s.Logf("successfully unjail with options %v", opt)
}

func (s *IntegrationTestSuite) execFeeGrant(c *chain, valIdx int, granter, grantee, spendLimit string, opt ...flagOption) {
	opt = append(opt, withKeyValue(flagFrom, granter))
	opt = append(opt, withKeyValue(flagSpendLimit, spendLimit))
	opts := applyOptions(c, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("granting %s fee from %s on chain %s", grantee, granter, c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		feegrant.ModuleName,
		"grant",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

func (s *IntegrationTestSuite) execFeeGrantRevoke(c *chain, valIdx int, granter, grantee string, opt ...flagOption) {
	opt = append(opt, withKeyValue(flagFrom, granter))
	opts := applyOptions(c, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("revoking %s fee grant from %s on chain %s", grantee, granter, c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		feegrant.ModuleName,
		"revoke",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

func (s *IntegrationTestSuite) execBankSend(
	c *chain,
	valIdx int,
	from,
	to,
	amt,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyOptions(c, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		banktypes.ModuleName,
		"send",
		from,
		to,
		amt,
		"-y",
	}
	for flag, value := range opts {
		chainCommand = append(chainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

type txBankSend struct {
	from      string
	to        string
	amt       string
	fees      string
	log       string
	expectErr bool
}

func (s *IntegrationTestSuite) execBankSendBatch(
	c *chain,
	valIdx int, //nolint:unparam
	txs ...txBankSend,
) int {
	sucessBankSendCount := 0

	for i := range txs {
		s.Logf(txs[i].log)

		s.execBankSend(c, valIdx, txs[i].from, txs[i].to, txs[i].amt, txs[i].fees, txs[i].expectErr)
		if !txs[i].expectErr {
			if !txs[i].expectErr {
				sucessBankSendCount++
			}
		}
	}

	return sucessBankSendCount
}

func (s *IntegrationTestSuite) execWithdrawAllRewards(c *chain, valIdx int, payee, fees string, expectErr bool) { //nolint:unparam
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		distributiontypes.ModuleName,
		"withdraw-all-rewards",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, payee),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

func (s *IntegrationTestSuite) execDistributionFundCommunityPool(c *chain, valIdx int, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Executing gaiad tx distribution fund-community-pool on chain %s", c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		distributiontypes.ModuleName,
		"fund-community-pool",
		amt,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("Successfully funded community pool")
}

func (s *IntegrationTestSuite) runGovExec(c *chain, valIdx int, submitterAddr, govCommand string, proposalFlags []string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		govtypes.ModuleName,
		govCommand,
	}

	generalFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	chainCommand = concatFlags(chainCommand, proposalFlags, generalFlags)

	s.Logf("Executing gaiad tx gov %s on chain %s", govCommand, c.ID)
	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("Successfully executed %s", govCommand)
}

func (s *IntegrationTestSuite) executeGKeysAddCommand(c *chain, valIdx int, name string, home string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	chainCommand := []string{
		c.ChainNodeBinary,
		keysCommand,
		"add",
		name,
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--keyring-backend=test",
		"--output=json",
	}

	var addrRecord AddressResponse
	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		// Gaiad keys add by default returns payload to stdErr
		if err := json.Unmarshal(stdErr, &addrRecord); err != nil {
			return false
		}
		return strings.Contains(addrRecord.Address, "cosmos")
	})
	return addrRecord.Address
}

func (s *IntegrationTestSuite) executeKeysList(c *chain, valIdx int, home string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	chainCommand := []string{
		c.ChainNodeBinary,
		keysCommand,
		"list",
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, func([]byte, []byte) bool {
		return true
	})
}

func (s *IntegrationTestSuite) executeDelegate(c *chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) { //nolint:unparam
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Executing gaiad tx staking delegate %s", c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		stakingtypes.ModuleName,
		"delegate",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) executeRedelegate(c *chain, valIdx int, amount, originalValOperAddress,
	newValOperAddress, delegatorAddr, home, delegateFees string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Executing gaiad tx staking redelegate %s", c.ID)

	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		stakingtypes.ModuleName,
		"redelegate",
		originalValOperAddress,
		newValOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}

func (s *IntegrationTestSuite) getLatestBlockHeight(c *chain, valIdx int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	type syncInfo struct {
		SyncInfo struct {
			LatestHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var currentHeight int
	chainCommand := []string{c.ChainNodeBinary, "status"}
	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		var (
			err   error
			block syncInfo
		)
		s.Require().NoError(json.Unmarshal(stdErr, &block))
		currentHeight, err = strconv.Atoi(block.SyncInfo.LatestHeight)
		s.Require().NoError(err)
		return currentHeight > 0
	})
	return currentHeight
}

func (s *IntegrationTestSuite) execSetWithdrawAddress(
	c *chain,
	valIdx int,
	fees,
	delegatorAddress,
	newWithdrawalAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Setting distribution withdrawal address on chain %s for %s to %s", c.ID, delegatorAddress, newWithdrawalAddress)
	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		distributiontypes.ModuleName,
		"set-withdraw-addr",
		newWithdrawalAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("Successfully set new distribution withdrawal address for %s to %s", delegatorAddress, newWithdrawalAddress)
}

func (s *IntegrationTestSuite) execWithdrawReward(
	c *chain,
	valIdx int,
	delegatorAddress,
	validatorAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.Logf("Withdrawing distribution rewards on chain %s for delegator %s from %s validator", c.ID, delegatorAddress, validatorAddress)
	chainCommand := []string{
		c.ChainNodeBinary,
		txCommand,
		distributiontypes.ModuleName,
		"withdraw-rewards",
		validatorAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300uatom"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, c, chainCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.Logf("Successfully withdrew distribution rewards for delegator %s from validator %s", delegatorAddress, validatorAddress)
}

func (s *IntegrationTestSuite) executeCeliniumTxCommand(ctx context.Context, c *chain, chainCommand []string, valIdx int, validation func([]byte, []byte) bool) {
	if validation == nil {
		validation = s.defaultExecValidation(c, 0)
	}
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.ID][valIdx].Container.ID,
		User:         "nonroot",
		Cmd:          chainCommand,
	})
	s.Require().NoError(err)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoError(err)

	stdOut := outBuf.Bytes()
	stdErr := errBuf.Bytes()
	if !validation(stdOut, stdErr) {
		s.Require().FailNowf("Exec validation failed", "stdout: %s, stderr: %s",
			string(stdOut), string(stdErr))
	}
}

func (s *IntegrationTestSuite) expectErrExecValidation(chain *chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		gotErr := chain.encfg.Codec.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			s.Require().True(expectErr)
		}

		endpoint := fmt.Sprintf("http://%s", s.valResources[chain.ID][valIdx].GetHostPort("1317/tcp"))
		// wait for the tx to be committed on chain
		s.Require().Eventuallyf(
			func() bool {
				gotErr := queryChainTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s",
			string(stdOut), string(stdErr),
		)
		return true
	}
}

func (s *IntegrationTestSuite) defaultExecValidation(chain *chain, valIdx int) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		if err := chain.encfg.Codec.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.String(), "code: 0") || txResp.Code == 0 {
			endpoint := fmt.Sprintf("http://%s", s.valResources[chain.ID][valIdx].GetHostPort("1317/tcp"))
			s.Require().Eventually(
				func() bool {
					return queryChainTx(endpoint, txResp.TxHash) == nil
				},
				time.Minute,
				5*time.Second,
				"stdOut: %s, stdErr: %s",
				string(stdOut), string(stdErr),
			)
			return true
		}
		return false
	}
}

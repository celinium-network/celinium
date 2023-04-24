package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type ForwardMetadata struct {
	Receiver string `json:"receiver"`
	Port     string `json:"port"`
	Channel  string `json:"channel"`
	// Timeout        time.Duration `json:"timeout"`
	// Retries        *uint8        `json:"retries,omitempty"`
	// Next           *string       `json:"next,omitempty"`
	// RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

func (s *IntegrationTestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container...")

	tmpDir, err := os.MkdirTemp("", "gaia-e2e-testnet-relayer-")
	s.Require().NoError(err)
	s.tmpDirs = append(s.tmpDirs, tmpDir)

	srcVal := s.srcChain.validators[0]
	ctlVal := s.ctlChain.validators[0]

	srcRly := s.srcChain.genesisAccounts[relayerAccountIndex]
	ctlRly := s.ctlChain.genesisAccounts[relayerAccountIndex]

	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = copyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.relayerResource, err = s.dkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", s.srcChain.ID, s.ctlChain.ID),
			Repository: "relayer",
			// Tag:        "1.0.0",
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("CELI_SRC_E2E_CHAIN_ID=%s", s.srcChain.ID),
				fmt.Sprintf("CELI_CTL_E2E_CHAIN_ID=%s", s.ctlChain.ID),
				fmt.Sprintf("CELI_SRC_E2E_VAL_MNEMONIC=%s", srcVal.mnemonic),
				fmt.Sprintf("CELI_CTL_E2E_VAL_MNEMONIC=%s", ctlVal.mnemonic),
				fmt.Sprintf("CELI_SRC_E2E_RLY_MNEMONIC=%s", srcRly.mnemonic),
				fmt.Sprintf("CELI_CTL_E2E_RLY_MNEMONIC=%s", ctlRly.mnemonic),
				fmt.Sprintf("CELI_SRC_E2E_VAL_HOST=%s", s.valResources[s.srcChain.ID][0].Container.Name[1:]),
				fmt.Sprintf("CELI_CTL_E2E_VAL_HOST=%s", s.valResources[s.ctlChain.ID][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s/state", s.relayerResource.GetHostPort("3031/tcp"))
	s.Require().Eventually(
		func() bool {
			resp, err := http.Get(endpoint) //nolint:gosec // this is a test
			if err != nil {
				return false
			}

			defer resp.Body.Close()

			bz, err := io.ReadAll(resp.Body)
			if err != nil {
				return false
			}

			var respBody map[string]interface{}
			if err := json.Unmarshal(bz, &respBody); err != nil {
				return false
			}

			status := respBody["status"].(string)
			result := respBody["result"].(map[string]interface{})

			return status == "success" && len(result["chains"].([]interface{})) == 2
		},
		5*time.Minute,
		time.Second,
		"hermes relayer not healthy",
	)

	s.T().Logf("started Hermes relayer container: %s", s.relayerResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Celinium chains
	s.createConnection()
	time.Sleep(10 * time.Second)
	s.createChannel()
}

func (s *IntegrationTestSuite) sendIBC(c *chain, valIdx int, sender, recipient, token, fees, note string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		c.ChainNodeBinary,
		txCommand,
		"ibc-transfer",
		"transfer",
		"transfer",
		"channel-0",
		recipient,
		token,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, s.srcChain.ID, sender, s.ctlChain.ID, recipient, note)
	s.executeGaiaTxCommand(ctx, c, ibcCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) createConnection() {
	s.T().Logf("connecting %s and %s chains via IBC", s.srcChain.ID, s.ctlChain.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.relayerResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"connection",
			"--a-chain",
			s.srcChain.ID,
			"--b-chain",
			s.ctlChain.ID,
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.srcChain.ID, s.ctlChain.ID)
}

func (s *IntegrationTestSuite) createChannel() {
	s.T().Logf("connecting %s and %s chains via IBC", s.srcChain.ID, s.ctlChain.ID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.relayerResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			txCommand,
			"chan-open-init",
			"--dst-chain",
			s.srcChain.ID,
			"--src-chain",
			s.ctlChain.ID,
			"--dst-connection",
			"connection-0",
			"--src-port=transfer",
			"--dst-port=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.srcChain.ID, s.ctlChain.ID)
}

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	time.Sleep(30 * time.Second)
	s.Run("send_celi_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances      sdk.Coins
			err           error
			beforeBalance int64
			ibcStakeDenom string
		)

		address, _ := s.srcChain.validators[0].keyRecord.GetAddress()
		sender := address.String()

		address, _ = s.ctlChain.validators[0].keyRecord.GetAddress()
		recipient := address.String()

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
		cdc := s.srcChain.encfg.Codec

		s.Require().Eventually(
			func() bool {
				balances, err = queryAllBalances(cdc, chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				beforeBalance = c.Amount.Int64()
				break
			}
		}

		tokenAmt := 3300000000
		fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)
		s.sendIBC(s.srcChain, 0, sender, recipient, strconv.Itoa(tokenAmt)+s.srcChain.Denom, fee.String(), "")

		s.Require().Eventually(
			func() bool {
				balances, err = queryAllBalances(cdc, chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((int64(tokenAmt) + beforeBalance), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

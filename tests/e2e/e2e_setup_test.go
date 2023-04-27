package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmjson "github.com/tendermint/tendermint/libs/json"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const (
	relayerAccountIndex = 0
	txCommand           = "tx"
	numberOfEvidences   = 10
	gas                 = 200000
	minGasPrice         = "0.00001"
	celiniumHomePath    = "/home/nonroot/.celinium"
	keysCommand         = "keys"
)

var (
	stakingAmount     = sdk.NewInt(200000000000)
	standardFeeAmount = sdk.NewInt(330000)
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs         []string
	srcChain        *chain
	ctlChain        *chain
	dkrPool         *dockertest.Pool
	dkrNet          *dockertest.Network
	relayerResource *dockertest.Resource
	valResources    map[string][]*dockertest.Resource
}

type AddressResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.Logf("setting up e2e integration test suite...")

	var err error
	s.srcChain, err = newChain("source")
	s.Require().NoError(err)

	s.ctlChain, err = newChain("control")
	s.Require().NoError(err)

	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-%s-testnet", s.srcChain.ID, s.ctlChain.ID))
	s.Require().NoError(err)

	s.valResources = make(map[string][]*dockertest.Resource)

	vestingMnemonic, err := createMnemonic()
	s.Require().NoError(err)

	jailedValMnemonic, err := createMnemonic()
	s.Require().NoError(err)

	// The boostrapping phase is as follows:
	//
	// 1. Initialize Celinium validator nodes.
	// 2. Create and initialize Celinium validator genesis files (both chains)
	// 3. Start both networks.
	// 4. Create and run IBC relayer containers.

	s.Logf("starting e2e infrastructure for chain A; chain-id: %s; datadir: %s", s.srcChain.ID, s.srcChain.dataDir)
	s.initNodes(s.srcChain)
	s.initGenesis(s.srcChain, vestingMnemonic, jailedValMnemonic)
	s.initValidatorConfigs(s.srcChain)
	s.runValidators(s.srcChain, 0)

	s.Logf("starting e2e infrastructure for chain B; chain-id: %s; datadir: %s", s.ctlChain.ID, s.ctlChain.dataDir)
	s.initNodes(s.ctlChain)
	s.initGenesis(s.ctlChain, vestingMnemonic, jailedValMnemonic)
	s.initValidatorConfigs(s.ctlChain)
	s.runValidators(s.ctlChain, 10)

	time.Sleep(10 * time.Second)
	s.runIBCRelayer()
}

func (s *IntegrationTestSuite) initNodes(c *chain) {
	s.Require().NoError(c.createAndInitValidators(2))
	/* Adding 4 accounts to val0 local directory
	c.genesisAccounts[0]: Relayer Wallet
	c.genesisAccounts[1]: ICA Owner
	c.genesisAccounts[2]: Test Account 1
	c.genesisAccounts[3]: Test Account 2
	*/
	s.Require().NoError(c.addAccountFromMnemonic(4))
	// Initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	var addrAll []sdk.AccAddress
	for _, val := range c.validators {
		address, _ := val.keyRecord.GetAddress()
		addrAll = append(addrAll, address)
	}

	for _, addr := range c.genesisAccounts {
		acctAddr, _ := addr.keyRecord.GetAddress()
		addrAll = append(addrAll, acctAddr)
	}

	s.Require().NoError(
		modifyGenesis(c.encfg.Codec, val0ConfigDir, "", c.InitBalance, addrAll, c.Denom),
	)
	// copy the genesis file to the remaining validators
	for _, val := range c.validators[1:] {
		_, err := copyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) initGenesis(c *chain, vestingMnemonic, jailedValMnemonic string) {
	var (
		serverCtx = server.NewDefaultContext()
		config    = serverCtx.Config
		validator = c.validators[0]
	)

	config.SetRoot(validator.configDir())
	config.Moniker = validator.moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var evidenceGenState evidencetypes.GenesisState
	s.Require().NoError(c.encfg.Codec.UnmarshalJSON(appGenState[evidencetypes.ModuleName], &evidenceGenState))

	evidenceGenState.Evidence = make([]*codectypes.Any, 10)
	for i := range evidenceGenState.Evidence {
		pk := ed25519.GenPrivKey()
		evidence := &evidencetypes.Equivocation{
			Height:           1,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
		}
		evidenceGenState.Evidence[i], err = codectypes.NewAnyWithValue(evidence)
		s.Require().NoError(err)
	}

	appGenState[evidencetypes.ModuleName], err = c.encfg.Codec.MarshalJSON(&evidenceGenState)
	s.Require().NoError(err)

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(c.encfg.Codec.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.validators))
	for i, val := range c.validators {
		// the first validator vote power >= 2/3 * totolPower
		stAmt := stakingAmount
		if i != 0 {
			stAmt = stakingAmount.QuoRaw(2).SubRaw(1)
		}
		createValmsg, err := val.buildCreateValidatorMsg(sdk.NewCoin(c.Denom, stAmt))
		s.Require().NoError(err)
		signedTx, err := val.signMsg(createValmsg)

		s.Require().NoError(err)

		txRaw, err := c.encfg.Codec.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	appGenState[genutiltypes.ModuleName], err = c.encfg.Codec.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)

	genDoc.AppState, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	bz, err := tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// vestingPeriod, err := generateVestingPeriod()
	// s.Require().NoError(err)

	rawTx, _, err := buildRawTx(c)
	s.Require().NoError(err)

	// write the updated genesis file to each validator.
	for _, val := range c.validators {
		err = writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
		s.Require().NoError(err)

		err = writeFile(filepath.Join(val.configDir(), rawTxFile), rawTx)
		s.Require().NoError(err)
	}
}

// initValidatorConfigs initializes the validator configs for the given chain.
func (s *IntegrationTestSuite) initValidatorConfigs(c *chain) {
	for i, val := range c.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()

		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.validators); j++ {
			if i == j {
				continue
			}

			peer := c.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")
		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrice, c.Denom)

		customAppTemplate := srvconfig.DefaultConfigTemplate
		srvconfig.SetConfigTemplate(customAppTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

// runValidators runs the validators in the chain
func (s *IntegrationTestSuite) runValidators(c *chain, portOffset int) {
	s.Logf("starting chain %s validator containers...", c.ID)

	s.valResources[c.ID] = make([]*dockertest.Resource, len(c.validators))
	for i, val := range c.validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.instanceName(),
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:%s", val.configDir(), celiniumHomePath),
			},
			Repository: c.DokcerImage,
		}

		s.Require().NoError(exec.Command("chmod", "-R", "0777", val.configDir()).Run()) //nolint:gosec // this is a test

		// expose the first validator for debugging and communication
		if val.index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
				"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
				"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
				"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
				"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
				"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
				"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
				"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
				"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
				"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
			}
		}

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[c.ID][i] = resource
		s.Logf("started Celinium %s validator container: %s", c.ID, resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}

			return true
		},
		5*time.Minute,
		time.Second,
		"chain node failed to produce blocks",
	)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("GAIA_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.Logf("tearing down e2e integration test suite...")

	s.Require().NoError(s.dkrPool.Purge(s.relayerResource))

	for _, vr := range s.valResources {
		for _, r := range vr {
			s.Require().NoError(s.dkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	os.RemoveAll(s.srcChain.dataDir)
	os.RemoveAll(s.ctlChain.dataDir)

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) Logf(format string, args ...any) {
	s.T().Log(time.Now().Format(time.RFC3339) + ": " + fmt.Sprintf(format, args...))
}

package e2e

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"

	celiniumapp "github.com/celinium-network/celinium/app"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

type encodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

type chain struct {
	dataDir             string
	ID                  string
	validators          []*validator
	accounts            []*account //nolint:unused
	genesisAccounts     []*account
	encfg               encodingConfig
	InitBalance         string
	Denom               string
	ChainNodeBinary     string
	ModuleBasicsGenesis func() (json.RawMessage, error)
	DokcerImage         string
}

func newChain(tag string) (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "celinium-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	encfg := celiniumapp.MakeEncodingConfig()
	return &chain{
		ID:      "celin-chain-" + tag + "-" + tmrand.Str(6),
		dataDir: tmpDir,
		encfg: encodingConfig{
			InterfaceRegistry: encfg.InterfaceRegistry,
			Codec:             encfg.Codec,
			TxConfig:          encfg.TxConfig,
			Amino:             encfg.Amino,
		},
		InitBalance:     "100000000000000000CELI",
		Denom:           "CELI",
		ChainNodeBinary: "celiniumd",
		ModuleBasicsGenesis: func() (json.RawMessage, error) {
			return json.MarshalIndent(celiniumapp.ModuleBasics.DefaultGenesis(encfg.Codec), "", " ")
		},
		DokcerImage: "celinium", // TODO Use the local docker image compiled by the current code instead
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.ID)
}

func (c *chain) createAndInitValidators(count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createAndInitValidatorsWithMnemonics(count int, mnemonics []string) error { //nolint:unused // this is called during e2e tests
	for i := 0; i < count; i++ {
		// create node
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKeyFromMnemonic("val", mnemonics[i]); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-gaia-%d", c.ID, index),
	}
}

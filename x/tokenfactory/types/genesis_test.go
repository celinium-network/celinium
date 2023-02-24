package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"celinium/x/tokenfactory/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "different admin from creator",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "no admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "moose",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "multiple denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/litecoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/demo1y6ejmxmq7d8l36pa4ytugvl6mrd9uy0wmvwzmf/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

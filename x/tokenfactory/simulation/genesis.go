package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"celinium/x/tokenfactory/types"
)

// RandomizedGenState generates a random GenesisState for tokenfactory
func RandomizedGenState(simState *module.SimulationState) {
	tfDefaultGen := types.DefaultGenesis()
	tfDefaultGen.Params.DenomCreationFee = sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10000000)))
	tfDefaultGenJson := simState.Cdc.MustMarshalJSON(tfDefaultGen)
	simState.GenState[types.ModuleName] = tfDefaultGenJson
}

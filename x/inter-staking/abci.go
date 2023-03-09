package interstaking

import (
	"time"

	"celinium/x/inter-staking/keeper"
	"celinium/x/inter-staking/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EndBlock(ctx sdk.Context, k *keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.ProcessDelegationTask(ctx)

	return nil
}

func BeginBlock(ctx sdk.Context, k *keeper.Keeper) {}

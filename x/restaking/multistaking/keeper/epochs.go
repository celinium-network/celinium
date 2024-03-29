package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

func (h Hooks) AfterEpochEnd(ctx sdk.Context, _ string, _ int64) {
}

// BeforeEpochStart implements types.EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	switch epochIdentifier {
	case types.RefreshAgentDelegationEpochID:
		h.k.RefreshAgentDelegationAmount(ctx)
	case types.CollectAgentStakingRewardEpochID:
		// TODO remove it from epoch ?
		h.k.CollectAgentsReward(ctx)
	}
}

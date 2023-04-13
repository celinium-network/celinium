package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

type Hooks struct {
	k Keeper
}

// AfterEpochEnd implements types.EpochHooks
func (Hooks) AfterEpochEnd(_ sdk.Context, _ string, _ int64) {
}

// BeforeEpochStart implements types.EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	switch epochIdentifier {
	case types.DelegationEpochIdentifier:
		h.k.CreateDelegationRecordForEpoch(ctx, epochNumber)

		delegationRecords := h.k.GetAllDelegationRecord(ctx)
		h.k.ProcessDelegationRecord(ctx, uint64(epochNumber), delegationRecords)

		h.k.UpdateRedeemRatio(ctx, delegationRecords)
	case types.UndelegationEpochIdentifier:
		h.k.CreateEpochUnbondings(ctx, epochNumber)

		h.k.ProcessUnbondings(ctx, uint64(epochNumber))
	case types.ReinvestEpochIdentifier:
		h.k.StartReInvest(ctx)
	default:
	}
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

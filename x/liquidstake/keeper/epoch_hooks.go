package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

// Hooks wrapper struct for liquid stake keeper
type Hooks struct {
	k Keeper
}

// AfterEpochEnd implements types.EpochHooks
func (Hooks) AfterEpochEnd(_ sdk.Context, _ string, _ int64) {
}

// BeforeEpochStart implements types.EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if epochIdentifier == types.DelegationEpochIdentifier {
		// Create new delegation for current epoch
		h.k.CreateDelegationRecordForEpoch(ctx, epochNumber)

		delegationRecords := h.k.GetAllDelegationRecord(ctx)

		h.k.ProcessDelegationRecord(ctx, uint64(epochNumber), delegationRecords)

		// update rate,

		// reinvest, start from a interchain query, maybe submit by offchain timer service?
	} else if epochIdentifier == types.UndelegationEpochIdentifier {
		// Create new unbondings for current epoch
		h.k.CreateEpochUnbondings(ctx, epochNumber)

		// Process Unbound
		h.k.ProcessUnbondings(ctx, uint64(epochNumber))
	}
}

var _ epochstypes.EpochHooks = Hooks{}

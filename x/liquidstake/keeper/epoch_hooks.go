package keeper

import (
	"celinium/x/liquidstake/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "celinium/x/epochs/types"
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
		h.k.CreateDepositRecordForEpoch(ctx, epochNumber)

		delegationRecords := h.k.GetAllDelegationRecord(ctx)

		h.k.ProcessDelegationRecord(ctx, uint64(epochNumber), delegationRecords)

		// update rate

		// reinvest
	} else if epochIdentifier == types.UndelegationEpochIdentifier {
		// Create new delegation for current epoch

		// distribute unboud token

		// 
	}
}

var _ epochstypes.EpochHooks = Hooks{}

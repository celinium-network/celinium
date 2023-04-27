package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/celinium-network/celinium/app/params"
	epochstypes "github.com/celinium-network/celinium/x/epochs/types"
)

type Hooks struct {
	k Keeper
}

// AfterEpochEnd implements types.EpochHooks
func (Hooks) AfterEpochEnd(_ sdk.Context, _ string, _ int64) {
}

// BeforeEpochStart implements types.EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if epochNumber < 0 {
		return
	}

	uEpochNumber := uint64(epochNumber)

	switch epochIdentifier {
	case appparams.DelegationEpochIdentifier:
		h.k.CreateEpochProxyDelegation(ctx, uEpochNumber)

		proxyDelegations := h.k.GetAllProxyDelegation(ctx)
		h.k.ProcessProxyDelegation(ctx, uEpochNumber, proxyDelegations)

		h.k.UpdateRedeemRatio(ctx, proxyDelegations)
	case appparams.UndelegationEpochIdentifier:
		h.k.CreateEpochUnbondings(ctx, uEpochNumber)

		h.k.ProcessUnbondings(ctx, uEpochNumber)
	case appparams.ReinvestEpochIdentifier:
		h.k.SetDistriWithdrawAddress(ctx)

		h.k.Reinvest(ctx)
	default:
	}
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

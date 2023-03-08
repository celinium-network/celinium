package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
)

// SourceChainAvaiable whether SourceChain is available.
func (k Keeper) SourceChainAvaiable(ctx sdk.Context, connectionID string, icaCtladdr string) bool {
	portID, err := icatypes.NewControllerPortID(icaCtladdr)
	if err != nil {
		return false
	}
	_, found := k.icaControllerKeeper.GetOpenActiveChannel(ctx, connectionID, portID)
	return found
}

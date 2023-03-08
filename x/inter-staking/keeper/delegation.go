package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/inter-staking/types"
)

func (k Keeper) SendCoinsFromDelegatorToICA(ctx sdk.Context, delegatorAddr string, icaCtlAddr string, coins sdk.Coins) error {
	delegatorAccount := sdk.MustAccAddressFromBech32(delegatorAddr)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAccount, types.ModuleName, coins); err != nil {
		return err
	}

	icaCtlAccount := sdk.MustAccAddressFromBech32(icaCtlAddr)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, icaCtlAccount, coins); err != nil {
		return err
	}

	return nil
}

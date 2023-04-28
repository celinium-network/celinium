package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

// UpdateRedeemRate update redeemrate for each source chain
// TODO Record the rate in the last few epochs, and then average?
func (k Keeper) UpdateRedeemRate(ctx sdk.Context, delegations []types.ProxyDelegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	chainProcessingAmts := k.GetDelegaionProcessingAmount(delegations)

	for ; iterator.Valid(); iterator.Next() {
		sourcechain := &types.SourceChain{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		// If the status of a ProxyDelegation is `ProxyDelegationDone`, it maybe remove from the store.
		// So the doneAmount = stakedAmount.
		doneAmount := sourcechain.StakedAmount
		processingAmount, found := chainProcessingAmts[sourcechain.ChainID]

		if (!found || processingAmount.IsZero()) || (doneAmount.IsNil() || doneAmount.IsZero()) {
			continue
		}

		derivationAmount := k.bankKeeper.GetSupply(ctx, sourcechain.DerivativeDenom)

		if derivationAmount.IsZero() {
			continue
		}

		sourcechain.Redemptionratio = sdk.NewDecFromInt(processingAmount).Add(sdk.NewDecFromInt(doneAmount)).
			Quo(sdk.NewDecFromInt(derivationAmount.Amount))

		k.SetSourceChain(ctx, sourcechain)
	}
}

func (k Keeper) GetDelegaionProcessingAmount(delegations []types.ProxyDelegation) map[string]math.Int {
	chainAmts := make(map[string]math.Int)
	for _, delegation := range delegations {
		if !types.IsProxyDelegationProcessing(delegation.Status) {
			continue
		}
		if delegation.Coin.Amount.IsZero() {
			continue
		}

		amt, found := chainAmts[delegation.ChainID]

		userDelegationAmt := delegation.Coin.Amount
		if !delegation.ReinvestAmount.IsZero() {
			userDelegationAmt = userDelegationAmt.Sub(delegation.ReinvestAmount)
		}

		if !found {
			chainAmts[delegation.ChainID] = userDelegationAmt
		} else {
			chainAmts[delegation.ChainID] = userDelegationAmt.Add(amt)
		}
	}
	return chainAmts
}

// ClaimUnbonding implement delegator claim reward and stake token.
func (k Keeper) ClaimUnbonding(ctx sdk.Context, deletator sdk.AccAddress, epoch uint64, chainID string) (math.Int, error) {
	undelegationRecord, found := k.GetUserUnbonding(ctx, chainID, epoch, deletator.String())
	if !found {
		return math.ZeroInt(), sdkerrors.Wrapf(types.ErrUserUndelegationNotExist, "chainID %s, epoch %d, address %s",
			chainID, epoch, deletator.String())
	}

	if undelegationRecord.CliamStatus != types.UserUnbondingClaimable {
		return math.ZeroInt(), sdkerrors.Wrapf(types.ErrUserUndelegationWatting, "chainID %s, epoch %d, address %s",
			chainID, epoch, deletator.String())
	}

	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return math.ZeroInt(), sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID %s", chainID)
	}

	chainDelegatorAccAddress, err := sdk.AccAddressFromBech32(sourceChain.DelegateAddress)
	if err != nil {
		return math.ZeroInt(), err
	}

	err = k.sendCoinsFromAccountToAccount(ctx, chainDelegatorAccAddress, deletator, sdk.NewCoins(undelegationRecord.RedeemCoin))
	if err != nil {
		return math.ZeroInt(), err
	}

	undelegationRecord.CliamStatus = types.UserUnbondingComplete

	k.SetUserUnbonding(ctx, undelegationRecord)
	return math.ZeroInt(), nil
}

package keeper

import (
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

// UpdateRedeemRatio update redeemrate for each source chain
// TODO Record the rate in the last few epochs, and then average?
// TODO make iterator sourcechain into function, it already used in `CreateDelegationRecordForEpoch`
func (k Keeper) UpdateRedeemRatio(ctx sdk.Context, records []types.DelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	for ; iterator.Valid(); iterator.Next() {
		sourcechain := &types.SourceChain{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		processingAmount := k.GetProcessingFundsFromRecords(sourcechain, records)
		doneAmount := sourcechain.StakedAmount

		if processingAmount.IsZero() && doneAmount.IsZero() {
			continue
		}

		derivationAmount := k.bankKeeper.GetSupply(ctx, sourcechain.DerivativeDenom)

		if derivationAmount.IsZero() {
			continue
		}

		sourcechain.Redemptionratio = sdk.NewDecFromInt(processingAmount).Add(sdk.NewDecFromInt(doneAmount)).Quo(sdk.NewDecFromInt(derivationAmount.Amount))

		k.SetSourceChain(ctx, sourcechain)
	}
}

// TODO, make it return `map[chainID]math.Int`, then just loop it once.
func (k Keeper) GetProcessingFundsFromRecords(sourceChain *types.SourceChain, records []types.DelegationRecord) math.Int {
	amount := math.ZeroInt()
	for _, record := range records {
		if record.ChainID != sourceChain.ChainID {
			continue
		}

		if !types.IsDelegationRecordProcessing(record.Status) {
			continue
		}
		if record.DelegationCoin.Amount.IsZero() {
			continue
		}

		amount = amount.Add(record.DelegationCoin.Amount)
	}
	return amount
}

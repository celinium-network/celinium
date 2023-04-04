package keeper

import (
	"celinium/x/liquidstake/types"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Delegate performs a liquid stake delegation. delegator transfer the ibcToken to module account then
// get derivative token by the rate.
func (k *Keeper) Delegate(ctx sdk.Context, chainID string, amount math.Int, delegator sdk.AccAddress) error {
	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", chainID)
	}

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, types.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", types.DelegationEpochIdentifier)
	}

	currentEpoch := uint64(epochInfo.CurrentEpoch)
	recordID, found := k.GetChianDelegationRecordID(ctx, chainID, currentEpoch)
	if !found {
		return sdkerrors.Wrapf(types.ErrNoExistDelegationRecord, "chainID %s, epoch %d", chainID, currentEpoch)
	}

	record, found := k.GetDelegationRecord(ctx, recordID)
	if !found {
		return sdkerrors.Wrapf(types.ErrNoExistDelegationRecord, "chainID %s, epoch %d, recorID %d", chainID, currentEpoch, recordID)
	}

	delegatorAccAddress := sdk.MustAccAddressFromBech32(sourceChain.DelegateAddress)
	// transfer ibc token to sourcechain's delegation account
	if err := k.sendCoinsFromAccountToAccount(ctx, delegator, delegatorAccAddress, sdk.Coins{sdk.NewCoin(sourceChain.IbcDenom, amount)}); err != nil {
		return err
	}

	// todo! TruncateInt calculations can be huge precision error
	derivativeCoinAmount := amount.Mul(sourceChain.Redemptionratio.TruncateInt())
	if err := k.mintCoins(ctx, delegatorAccAddress, sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, derivativeCoinAmount)}); err != nil {
		return err
	}

	record.DelegationCoin = record.DelegationCoin.AddAmount(amount)

	k.SetDelegationRecord(ctx, recordID, record)

	return nil
}

// BeginLiquidStake start liquid stake on source chain with provide delegation records.
// This process will continue to advance the status of the DelegationRecord according to the IBC ack.
func (k *Keeper) BeginLiquidStakeProcess(ctx sdk.Context, records []types.DelegationRecord) {
	for _, record := range records {
		if record.Status == types.Pending {
		}
	}
}

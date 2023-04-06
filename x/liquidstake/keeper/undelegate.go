package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/liquidstake/types"
)

func (k Keeper) Undelegate(ctx sdk.Context, chainID string, amount math.Int, delegator sdk.AccAddress /*,receiver sdk.AccAddress*/) error {
	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", chainID)
	}

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, types.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", types.DelegationEpochIdentifier)
	}

	// TODO, epoch should be uint64 or int64
	currentEpoch := uint64(epochInfo.CurrentEpoch)
	delegatorAddr := delegator.String()

	_, found = k.GetUndelegationRecord(ctx, chainID, currentEpoch, delegatorAddr)
	if found {
		return sdkerrors.Wrapf(types.ErrRepeatUndelegate, "epoch %d", currentEpoch)
	}

	// TODO, How to confirm the accuracy of calcualate ?
	receiveAmount := sdk.NewDecFromInt(amount).Mul(sourceChain.Redemptionratio).TruncateInt()
	if sourceChain.StakedAmount.LT(receiveAmount) {
		return sdkerrors.Wrapf(types.ErrInternalError, "undelegate too mach, max %s, get %s", sourceChain.StakedAmount, receiveAmount)
	}

	delegatorDerivativeTokenAmount := k.bankKeeper.GetBalance(ctx, delegator, sourceChain.DerivativeDenom)
	if delegatorDerivativeTokenAmount.Amount.LT(amount) {
		return sdkerrors.Wrapf(types.ErrInsufficientFunds, "burn %s, expectd: %s, own %s",
			sourceChain.DerivativeDenom,
			amount,
			delegatorDerivativeTokenAmount.Amount)
	}

	undelegationRecord := types.UndelegationRecord{
		ID:          types.AssembleUndelegationRecordID(chainID, currentEpoch, delegatorAddr),
		ChainID:     chainID,
		Epoch:       currentEpoch,
		Delegator:   delegatorAddr,
		Receiver:    "", // TODO unused, remove,
		RedeemToken: sdk.NewCoin(sourceChain.NativeDenom, receiveAmount),
		CliamStatus: types.UndelegationPending,
	}

	// update related Unbonding by chainID
	curEpochUnbondings, found := k.GetUnboundingsForEpoch(ctx, currentEpoch)
	if !found {
		return sdkerrors.Wrapf(types.ErrEpochUnbondingNotExist, "epoch %d", currentEpoch)
	}

	var curEpochSourceChainUnbonding types.Unbonding
	chainUnbondingIndex := -1
	for i, unbonding := range curEpochUnbondings.Unbondings {
		if unbonding.ChainID == chainID {
			curEpochSourceChainUnbonding = unbonding
			chainUnbondingIndex = i
		}
	}

	// unbonding of the chain is not created, then create it now.
	if chainUnbondingIndex == -1 {
		curEpochSourceChainUnbonding = types.Unbonding{
			ChainID:                chainID,
			BurnedDerivativeAmount: sdk.ZeroInt(),
			RedeemNativeToken:      sdk.NewCoin(sourceChain.NativeDenom, sdk.ZeroInt()),
			UnbondTIme:             0,
			Status:                 0,
			UserUnbondRecordIds:    []string{},
		}
	}

	curEpochSourceChainUnbonding.BurnedDerivativeAmount = curEpochSourceChainUnbonding.BurnedDerivativeAmount.Add(amount)
	curEpochSourceChainUnbonding.RedeemNativeToken = curEpochSourceChainUnbonding.RedeemNativeToken.AddAmount(receiveAmount)
	curEpochSourceChainUnbonding.UserUnbondRecordIds = append(curEpochSourceChainUnbonding.UserUnbondRecordIds, undelegationRecord.ChainID)

	if chainUnbondingIndex == -1 {
		// just append it
		curEpochUnbondings.Unbondings = append(curEpochUnbondings.Unbondings, curEpochSourceChainUnbonding)
	} else {
		// update with the index
		curEpochUnbondings.Unbondings[chainUnbondingIndex] = curEpochSourceChainUnbonding
	}

	k.SetUndelegationRecord(ctx, &undelegationRecord)

	k.SetUnboundingsForEpoch(ctx, curEpochUnbondings)

	return nil
}

func (k Keeper) GetUndelegationRecord(ctx sdk.Context, chainID string, epoch uint64, delegator string) (*types.UndelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.GetUndelegationRecordKey(chainID, epoch, delegator)))
	if bz == nil {
		return nil, false
	}

	record := types.UndelegationRecord{}
	k.cdc.MustUnmarshal(bz, &record)

	return &record, true
}

func (k Keeper) SetUndelegationRecord(ctx sdk.Context, undelegationRecord *types.UndelegationRecord) {
	store := ctx.KVStore(k.storeKey)

	key := types.GetUndelegationRecordKey(undelegationRecord.ChainID, undelegationRecord.Epoch, undelegationRecord.Delegator)
	bz := k.cdc.MustMarshal(undelegationRecord)
	store.Set([]byte(key), bz)
}

func (k Keeper) GetUnboundingsForEpoch(ctx sdk.Context, epoch uint64) (*types.EpochUnbondings, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetEpochUnbondingsKey(epoch))
	if bz == nil {
		return nil, false
	}

	unbondings := types.EpochUnbondings{}
	k.cdc.MustUnmarshal(bz, &unbondings)

	return &unbondings, true
}

func (k Keeper) SetUnboundingsForEpoch(ctx sdk.Context, unbondings *types.EpochUnbondings) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(unbondings)

	store.Set(types.GetEpochUnbondingsKey(unbondings.Epoch), bz)
}

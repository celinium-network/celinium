package keeper

import (
	"sort"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

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
	curEpochUnbondings, found := k.GetEpochUnboundings(ctx, currentEpoch)
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
	curEpochSourceChainUnbonding.UserUnbondRecordIds = append(curEpochSourceChainUnbonding.UserUnbondRecordIds, undelegationRecord.ID)

	if chainUnbondingIndex == -1 {
		// just append it
		curEpochUnbondings.Unbondings = append(curEpochUnbondings.Unbondings, curEpochSourceChainUnbonding)
	} else {
		// update with the index
		curEpochUnbondings.Unbondings[chainUnbondingIndex] = curEpochSourceChainUnbonding
	}

	k.SetUndelegationRecord(ctx, &undelegationRecord)

	k.SetEpochUnboundings(ctx, curEpochUnbondings)

	return nil
}

func (k Keeper) GetUndelegationRecord(ctx sdk.Context, chainID string, epoch uint64, delegator string) (*types.UndelegationRecord, bool) {
	id := types.AssembleUndelegationRecordID(chainID, epoch, delegator)

	return k.GetUndelegationRecordByID(ctx, id)
}

func (k Keeper) GetUndelegationRecordByID(ctx sdk.Context, ID string) (*types.UndelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(types.GetUndelegationRecordKeyFromID(ID)))
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

func (k Keeper) GetEpochUnboundings(ctx sdk.Context, epoch uint64) (*types.EpochUnbondings, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetEpochUnbondingsKey(epoch))
	if bz == nil {
		return nil, false
	}

	unbondings := types.EpochUnbondings{}
	k.cdc.MustUnmarshal(bz, &unbondings)

	return &unbondings, true
}

func (k Keeper) SetEpochUnboundings(ctx sdk.Context, unbondings *types.EpochUnbondings) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(unbondings)

	store.Set(types.GetEpochUnbondingsKey(unbondings.Epoch), bz)
}

// ProcessUnbondings advance the Unbondings in the past epoch into the next status
func (k Keeper) ProcessUnbondings(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EpochUnbondingsPrefix)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		if bz == nil {
			// TODO why come here ?
			continue
		}
		epochUnbondings := types.EpochUnbondings{}
		k.cdc.MustUnmarshal(bz, &epochUnbondings)

		// epoch not past
		if epochUnbondings.Epoch >= epochNumber {
			continue
		}
		err := k.ProcessEpochUnbondings(ctx, epochUnbondings.Epoch, epochUnbondings.Unbondings)
		if err != nil {
			k.SetEpochUnboundings(ctx, &epochUnbondings)
		}
	}
}

func (k Keeper) ProcessEpochUnbondings(ctx sdk.Context, epoch uint64, unbondings []types.Unbonding) error {
	pendingUnbondAmount := make(map[string]math.Int)
	sourceChainTemp := make(map[string]*types.SourceChain)
	completeUnbondAmmount := make(map[string]math.Int)

	chainIDs := make([]string, 0)

	// TODO use index loop style.
	for i, unbonding := range unbondings {
		sourceChian, found := k.GetSourceChain(ctx, unbonding.ChainID)
		if !found {
			// TODO why come here ?
			continue
		}

		if !k.sourceChainAvaiable(ctx, sourceChian) {
			continue
		}

		chainIDs = append(chainIDs, unbonding.ChainID)

		sourceChainTemp[unbonding.ChainID] = sourceChian

		switch unbonding.Status {
		case types.UnbondingPending:
			existAmount, ok := pendingUnbondAmount[unbonding.ChainID]
			if !ok {
				existAmount = sdk.ZeroInt()
			}
			pendingUnbondAmount[unbonding.ChainID] = existAmount.Add(unbonding.RedeemNativeToken.Amount)
			unbondings[i].Status = types.UnbondingStart

		case types.UnbondingTransferFailed:
			// TODO must retry?
		case types.UnbondingStartFailed:
			// TODO become pending and retry next epoch or retry now ?
			// retry now maybe deadloop ?
		case types.UnbondingWaitting:
			// TODO timestamp become int64
			if ctx.BlockTime().Before(time.Unix(int64(unbonding.UnbondTIme), 0).Add(5 * time.Minute)) {
				continue
			}

			existAmount, ok := completeUnbondAmmount[unbonding.ChainID]
			if !ok {
				existAmount = sdk.ZeroInt()
			}
			completeUnbondAmmount[unbonding.ChainID] = existAmount.Add(unbonding.RedeemNativeToken.Amount)
			unbondings[i].Status = types.UnbondingWithdraw
		default:
		}
	}

	sort.Strings(chainIDs)

	for _, chainID := range chainIDs {
		k.submitUnbondICATransaction(ctx, sourceChainTemp[chainID], pendingUnbondAmount[chainID], epoch)
	}

	for _, chainID := range chainIDs {
		k.submitWithdrawUnbondICATransaction(ctx, sourceChainTemp[chainID], completeUnbondAmmount[chainID], epoch)
	}

	return nil
}

func (k Keeper) submitUnbondICATransaction(ctx sdk.Context, sourceChain *types.SourceChain, amount math.Int, epoch uint64) error {
	validatorAllocateFunds := sourceChain.AllocateFundsForValidator(amount)

	// TODO, Ensuring the order of Validators seems to be easy, as long as the order is determined when modifying them.
	sort.Slice(sourceChain.Validators, func(i, j int) bool {
		return strings.Compare(sourceChain.Validators[i].Address, sourceChain.Validators[j].Address) >= 0
	})

	undelegateMsgs := make([]proto.Message, 0)

	sourceChainDelegateAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ChainID, sourceChain.DelegateAddress)
	sourceChainDelegateAddress := sdk.MustAccAddressFromBech32(sourceChainDelegateAddr)

	for _, v := range sourceChain.Validators {
		valAddress, err := sdk.ValAddressFromBech32(v.Address)
		if err != nil {
			return err
		}

		undelegateMsgs = append(undelegateMsgs, stakingtypes.NewMsgUndelegate(
			sourceChainDelegateAddress,
			valAddress,
			sdk.NewCoin(sourceChain.NativeDenom, math.NewIntFromBigInt(validatorAllocateFunds[v.Address])),
		))
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, undelegateMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	portID, err := icatypes.NewControllerPortID(sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	// TODO timeout ?
	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	sequence, err := k.icaCtlKeeper.SendTx(ctx, nil, sourceChain.ConnectionID, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	unbondCallArgs := types.UnbondCallbackArgs{
		Epoch:   epoch,
		ChainID: sourceChain.ChainID,
	}

	bzArgs := k.cdc.MustMarshal(&unbondCallArgs)

	callback := types.IBCCallback{
		CallType: types.UnbondCall,
		Args:     string(bzArgs),
	}

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) submitWithdrawUnbondICATransaction(ctx sdk.Context, sourceChain *types.SourceChain, amount math.Int, epoch uint64) error {
	validatorAllocateFunds := sourceChain.AllocateFundsForValidator(amount)

	// TODO, Ensuring the order of Validators seems to be easy, as long as the order is determined when modifying them.
	sort.Slice(sourceChain.Validators, func(i, j int) bool {
		return strings.Compare(sourceChain.Validators[i].Address, sourceChain.Validators[j].Address) >= 0
	})

	witdrawMsgs := make([]proto.Message, 0)

	sourceChainDelegateAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ChainID, sourceChain.DelegateAddress)

	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	for _, v := range sourceChain.Validators {
		witdrawMsgs = append(witdrawMsgs, transfertypes.NewMsgTransfer(
			transfertypes.PortID, // TODO the source chain maybe not use the default ibc transfer port. config it.
			sourceChain.TrasnferChannelID,
			sdk.NewCoin(sourceChain.NativeDenom, math.NewIntFromBigInt(validatorAllocateFunds[v.Address])),
			sourceChainDelegateAddr,
			sourceChain.UnboudAddress,
			ibcclienttypes.Height{},
			uint64(timeoutTimestamp),
			"",
		))
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, witdrawMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	portID, err := icatypes.NewControllerPortID(sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	// TODO timeout ?
	sequence, err := k.icaCtlKeeper.SendTx(ctx, nil, sourceChain.ConnectionID, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	unbondCallArgs := types.UnbondCallbackArgs{
		Epoch:   epoch,
		ChainID: sourceChain.ChainID,
	}

	bzArgs := k.cdc.MustMarshal(&unbondCallArgs)

	callback := types.IBCCallback{
		CallType: types.WithdrawUnbondCall,
		Args:     string(bzArgs),
	}

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}
package keeper

import (
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (k Keeper) Undelegate(ctx sdk.Context, chainID string, amount math.Int, delegator sdk.AccAddress) (*types.UserUnbonding, error) {
	sourceChain, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownSourceChain, "unknown source chain, chainID: %s", chainID)
	}

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, appparams.UndelegationEpochIdentifier)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch, epoch identifier: %s", appparams.UndelegationEpochIdentifier)
	}

	// save convert from int64 to uint64 , guaranteed by epoch handle entrypoint.
	currentEpoch := uint64(epochInfo.CurrentEpoch)
	delegatorAddr := delegator.String()

	_, found = k.GetUserUnbonding(ctx, chainID, currentEpoch, delegatorAddr)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrRepeatUndelegate, "epoch %d", currentEpoch)
	}

	receiveAmount := sdk.NewDecFromInt(amount).Mul(sourceChain.Redemptionratio).TruncateInt()
	if sourceChain.StakedAmount.LT(receiveAmount) {
		return nil, sdkerrors.Wrapf(types.ErrInternalError, "undelegate too mach, max %s, get %s", sourceChain.StakedAmount, receiveAmount)
	}

	// send coin from user to module account.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegator, types.ModuleName,
		sdk.Coins{sdk.NewCoin(sourceChain.DerivativeDenom, amount)}); err != nil {
		return nil, err
	}

	userUnbonding := types.UserUnbonding{
		ID:          types.AssembleUserUnbondingID(chainID, currentEpoch, delegatorAddr),
		ChainID:     chainID,
		Epoch:       currentEpoch,
		Delegator:   delegatorAddr,
		RedeemCoin:  sdk.NewCoin(sourceChain.IbcDenom, receiveAmount),
		CliamStatus: types.UserUnbondingPending,
	}

	// update related ProxyUnbonding by chainID
	curEpochProxyUnbondings, found := k.GetEpochProxyUnboundings(ctx, currentEpoch)
	if !found {
		curEpochProxyUnbondings = k.CreateProxyUnbondingForEpoch(ctx, currentEpoch)
	}

	var curEpochChainProxyUnbonding types.ProxyUnbonding
	chainProxyUnbondingIndex := -1
	for i, unbonding := range curEpochProxyUnbondings.Unbondings {
		if unbonding.ChainID == chainID {
			curEpochChainProxyUnbonding = unbonding
			chainProxyUnbondingIndex = i
		}
	}

	// ProxyUnbonding of the chain is not created, then create it now.
	if chainProxyUnbondingIndex == -1 {
		curEpochChainProxyUnbonding = types.ProxyUnbonding{
			ChainID:                chainID,
			BurnedDerivativeAmount: sdk.ZeroInt(),
			RedeemNativeToken:      sdk.NewCoin(sourceChain.NativeDenom, sdk.ZeroInt()),
			UnbondTime:             0,
			Status:                 0,
			UserUnbondingIds:       []string{},
		}
	}

	curEpochChainProxyUnbonding.BurnedDerivativeAmount = curEpochChainProxyUnbonding.BurnedDerivativeAmount.Add(amount)
	curEpochChainProxyUnbonding.RedeemNativeToken = curEpochChainProxyUnbonding.RedeemNativeToken.AddAmount(receiveAmount)
	curEpochChainProxyUnbonding.UserUnbondingIds = append(curEpochChainProxyUnbonding.UserUnbondingIds, userUnbonding.ID)

	if chainProxyUnbondingIndex == -1 {
		// just append it
		curEpochProxyUnbondings.Unbondings = append(curEpochProxyUnbondings.Unbondings, curEpochChainProxyUnbonding)
	} else {
		// replace it by the index
		curEpochProxyUnbondings.Unbondings[chainProxyUnbondingIndex] = curEpochChainProxyUnbonding
	}

	k.SetUserUnbonding(ctx, &userUnbonding)

	k.SetEpochProxyUnboundings(ctx, curEpochProxyUnbondings)

	return &userUnbonding, nil
}

func (k Keeper) GetUserUnbonding(ctx sdk.Context, chainID string, epoch uint64, delegator string) (*types.UserUnbonding, bool) {
	id := types.AssembleUserUnbondingID(chainID, epoch, delegator)

	return k.GetUserUnbondingID(ctx, id)
}

func (k Keeper) GetUserUnbondingID(ctx sdk.Context, id string) (*types.UserUnbonding, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(types.GetUndelegationRecordKeyFromID(id)))
	if bz == nil {
		return nil, false
	}

	record := types.UserUnbonding{}
	k.cdc.MustUnmarshal(bz, &record)

	return &record, true
}

func (k Keeper) SetUserUnbonding(ctx sdk.Context, userUnbonding *types.UserUnbonding) {
	store := ctx.KVStore(k.storeKey)

	key := types.GetUserUnbondingKey(userUnbonding.ChainID, userUnbonding.Epoch, userUnbonding.Delegator)
	bz := k.cdc.MustMarshal(userUnbonding)
	store.Set([]byte(key), bz)
}

func (k Keeper) GetEpochProxyUnboundings(ctx sdk.Context, epoch uint64) (*types.EpochProxyUnbonding, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetEpochUnbondingsKey(epoch))
	if bz == nil {
		return nil, false
	}

	unbondings := types.EpochProxyUnbonding{}
	k.cdc.MustUnmarshal(bz, &unbondings)

	return &unbondings, true
}

func (k Keeper) SetEpochProxyUnboundings(ctx sdk.Context, unbondings *types.EpochProxyUnbonding) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(unbondings)

	store.Set(types.GetEpochUnbondingsKey(unbondings.Epoch), bz)
}

// ProcessUndelegationEpoch advance the Unbondings in the past epoch into the next status
func (k Keeper) ProcessUndelegationEpoch(ctx sdk.Context, epochNumber uint64) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EpochUnbondingsPrefix)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		if bz == nil {
			continue
		}
		epochProxyUnbondings := types.EpochProxyUnbonding{}
		k.cdc.MustUnmarshal(bz, &epochProxyUnbondings)

		// epoch not past
		if epochProxyUnbondings.Epoch >= epochNumber {
			continue
		}
		if err := k.ProcessEpochProxyUnbondings(ctx, epochProxyUnbondings.Epoch, epochProxyUnbondings.Unbondings); err != nil {
			continue
		}
		// save the changed epochUnbondings
		k.SetEpochProxyUnboundings(ctx, &epochProxyUnbondings)
	}
}

func (k Keeper) ProcessEpochProxyUnbondings(ctx sdk.Context, epoch uint64, proxyUnbondings []types.ProxyUnbonding) error {
	pendingUnbondAmount := make(map[string]math.Int)
	sourceChainTemp := make(map[string]*types.SourceChain)
	completeUnbondAmmount := make(map[string]math.Int)

	chainIDs := make([]string, 0)

	for i, unbonding := range proxyUnbondings {
		sourceChian, found := k.GetSourceChain(ctx, unbonding.ChainID)
		if !found {
			continue
		}

		if !k.sourceChainAvaiable(ctx, sourceChian) {
			continue
		}

		chainIDs = append(chainIDs, unbonding.ChainID)

		sourceChainTemp[unbonding.ChainID] = sourceChian

		switch unbonding.Status {
		case types.ProxyUnbondingPending:
			existAmount, ok := pendingUnbondAmount[unbonding.ChainID]
			if !ok {
				existAmount = sdk.ZeroInt()
			}
			pendingUnbondAmount[unbonding.ChainID] = existAmount.Add(unbonding.RedeemNativeToken.Amount)
			proxyUnbondings[i].Status = types.ProxyUnbondingStart

		case types.ProxyUnbondingTransferFailed:
			// TODO must retry?
		case types.ProxyUnbondingStartFailed:
			// TODO become pending and retry next epoch or retry now ?
			// retry now maybe deadloop ?
		case types.ProxyUnbondingWaitting:
			if ctx.BlockTime().Before(time.Unix(0, int64(unbonding.UnbondTime)).Add(5 * time.Minute)) {
				continue
			}

			existAmount, ok := completeUnbondAmmount[unbonding.ChainID]
			if !ok {
				existAmount = sdk.ZeroInt()
			}
			completeUnbondAmmount[unbonding.ChainID] = existAmount.Add(unbonding.RedeemNativeToken.Amount)
			proxyUnbondings[i].Status = types.ProxyUnbondingWithdraw
		default:
		}
	}

	sort.Strings(chainIDs)

	for _, chainID := range chainIDs {
		amount, ok := pendingUnbondAmount[chainID]
		if !ok || amount.IsZero() {
			continue
		}
		k.undelegateOnSourceChain(ctx, sourceChainTemp[chainID], pendingUnbondAmount[chainID], epoch)
	}

	for _, chainID := range chainIDs {
		amount, ok := completeUnbondAmmount[chainID]
		if !ok || amount.IsZero() {
			continue
		}
		k.withdrawUnbondFromSourceChain(ctx, sourceChainTemp[chainID], completeUnbondAmmount[chainID], epoch)
	}

	return nil
}

func (k Keeper) undelegateOnSourceChain(ctx sdk.Context, sourceChain *types.SourceChain, amount math.Int, epoch uint64) error {
	allocVals := sourceChain.AllocateTokenForValidator(amount)

	undelegateMsgs := make([]proto.Message, 0)
	sourceChainUnbondAddress, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	for _, valFund := range allocVals.Validators {
		undelegateMsgs = append(undelegateMsgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: sourceChainUnbondAddress,
			ValidatorAddress: valFund.Address,
			Amount: sdk.Coin{
				Denom:  sourceChain.NativeDenom,
				Amount: valFund.TokenAmount,
			},
		})
	}

	sequence, portID, err := k.sendIBCMsg(ctx, undelegateMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	unbondCallArgs := types.UnbondCallbackArgs{
		Epoch:      epoch,
		ChainID:    sourceChain.ChainID,
		Validators: allocVals.Validators,
	}

	bzArgs := k.cdc.MustMarshal(&unbondCallArgs)

	callback := types.IBCCallback{
		CallType: types.UndelegateCall,
		Args:     string(bzArgs),
	}

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) withdrawUnbondFromSourceChain(ctx sdk.Context, sourceChain *types.SourceChain, amount math.Int, epoch uint64) error {
	sourceChainUnbondAddr, err := k.GetSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	if err != nil {
		return err
	}

	witdrawMsgs := make([]proto.Message, 0)
	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	allocVals := sourceChain.AllocateTokenForValidator(amount)

	for _, valFund := range allocVals.Validators {
		witdrawMsgs = append(witdrawMsgs, transfertypes.NewMsgTransfer(
			transfertypes.PortID, // TODO the source chain maybe not use the default ibc transfer port. config it.
			sourceChain.TransferChannelID,
			sdk.NewCoin(sourceChain.NativeDenom, valFund.TokenAmount),
			sourceChainUnbondAddr,
			sourceChain.DelegateAddress,
			ibcclienttypes.Height{},
			uint64(timeoutTimestamp),
			"",
		))
	}

	sequence, portID, err := k.sendIBCMsg(ctx, witdrawMsgs, sourceChain.ConnectionID, sourceChain.DelegateAddress)
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

package keeper

import (
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

func (k Keeper) MultiStakingDelegate(ctx sdk.Context, msg types.MsgMultiStakingDelegate) error {
	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)
	if strings.Compare(msg.Amount.Denom, defaultBondDenom) == 0 {
		return sdkerrors.Wrapf(types.ErrForbidStakingDenom, "denom: %s is native token", msg.Amount.Denom)
	}

	if !k.denomInWhiteList(ctx, msg.Amount.Denom) {
		return sdkerrors.Wrapf(types.ErrForbidStakingDenom, "denom: %s not in white list", msg.Amount.Denom)
	}

	agent := k.GetOrCreateMultiStakingAgent(ctx, msg.Amount.Denom, msg.ValidatorAddress)
	delegatorAccAddr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	if err := k.depositAndDelegate(ctx, agent, msg.Amount, delegatorAccAddr); err != nil {
		return err
	}

	shares := agent.CalculateShares(msg.Amount.Amount)
	agent.Shares = agent.Shares.Add(shares)
	agent.StakedAmount = agent.StakedAmount.Add(msg.Amount.Amount)

	k.SetMultiStakingAgent(ctx, agent)
	k.IncreaseMultiStakingShares(ctx, shares, agent.Id, msg.DelegatorAddress)

	return nil
}

func (k Keeper) depositAndDelegate(ctx sdk.Context, agent *types.MultiStakingAgent, amount sdk.Coin, delegator sdk.AccAddress) error {
	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.AgentDelegatorAddress)

	validator, err := k.agentValidator(ctx, agent)
	if err != nil {
		return err
	}

	if err := k.sendCoinsFromAccountToAccount(ctx, delegator, agentDelegateAccAddr, sdk.Coins{amount}); err != nil {
		return err
	}

	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)
	bondTokenAmt, err := k.EquivalentCoinCalculator(ctx, amount, defaultBondDenom)
	if err != nil {
		return err
	}
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{bondTokenAmt}); err != nil {
		return err
	}

	k.bankKeeper.AddSupplyOffset(ctx, defaultBondDenom, bondTokenAmt.Amount)

	if _, err = k.stakingkeeper.Delegate(ctx,
		agentDelegateAccAddr, bondTokenAmt.Amount,
		stakingtypes.Unbonded, *validator, true,
	); err != nil {
		return err
	}

	return nil
}

func (k Keeper) MultiStakingUndelegate(ctx sdk.Context, msg *types.MsgMultiStakingUndelegate) error {
	agent, found := k.GetMultiStakingAgent(ctx, msg.Amount.Denom, msg.ValidatorAddress)
	if found {
		return types.ErrNotExistedAgent
	}

	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)
	defaultAmt, err := k.EquivalentCoinCalculator(ctx, msg.Amount, defaultBondDenom)
	if err != nil {
		return err
	}

	delegatorAccAddr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	stakedShares, err := k.stakingkeeper.ValidateUnbondAmount(ctx, delegatorAccAddr, valAddr, defaultAmt.Amount)
	if err != nil {
		return nil
	}

	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.AgentDelegatorAddress)

	undelegationCoins, err := k.stakingkeeper.InstantUndelegate(ctx, agentDelegateAccAddr, valAddr, stakedShares)
	if err != nil {
		return nil
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		agentDelegateAccAddr, types.ModuleName, undelegationCoins,
	); err != nil {
		return nil
	}

	unbonding := k.GetOrCreateMultiStakingUnbonding(ctx, agent.Id, msg.DelegatorAddress)
	unbondingTime := k.stakingkeeper.GetParams(ctx).UnbondingTime

	// TODO Whether the length of Entry should be limited ?
	undelegateCompleteTime := ctx.BlockTime().Add(unbondingTime)
	unbonding.Entries = append(unbonding.Entries, types.MultiStakingUnbondingEntry{
		CompletionTime: undelegateCompleteTime,
		InitialBalance: undelegationCoins[0],
		Balance:        undelegationCoins[0],
	})

	k.SetMultiStakingUnbonding(ctx, agent.Id, msg.DelegatorAddress, unbonding)
	removeShares := agent.CalculateShares(msg.Amount.Amount)

	agent.Shares = agent.Shares.Sub(removeShares)
	agent.StakedAmount = agent.StakedAmount.Sub(msg.Amount.Amount)

	k.DecreaseMultiStakingShares(ctx, removeShares, agent.Id, msg.DelegatorAddress)
	k.SetMultiStakingAgent(ctx, agent)
	k.InsertUBDQueue(ctx, unbonding, undelegateCompleteTime)

	return nil
}

func (k Keeper) agentValidator(ctx sdk.Context, agent *types.MultiStakingAgent) (*stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(agent.AgentDelegatorAddress)
	if err != nil {
		return nil, err
	}

	validator, found := k.stakingkeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNotExistedValidator, "address %s", valAddr)
	}
	return &validator, nil
}

func (k Keeper) denomInWhiteList(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMultiStakingDenomWhiteList(ctx)
	if !found {
		return false
	}
	for _, wd := range whiteList.DenomList {
		if wd == denom {
			return true
		}
	}
	return false
}

func (k Keeper) GetOrCreateMultiStakingAgent(ctx sdk.Context, denom, valAddr string) *types.MultiStakingAgent {
	agent, found := k.GetMultiStakingAgent(ctx, denom, valAddr)
	if found {
		return agent
	}

	newAgentID := k.GetLatestMultiStakingAgentID(ctx)
	newAccount := k.GenerateAccount(ctx, denom, valAddr)

	agent = &types.MultiStakingAgent{
		Id:                    newAgentID,
		StakeDenom:            denom,
		AgentDelegatorAddress: newAccount.Address,
		WithdrawAddress:       newAccount.Address,
		StakedAmount:          math.ZeroInt(),
		RewardAmount:          math.ZeroInt(),
	}

	return agent
}

func (k Keeper) GenerateAccount(ctx sdk.Context, prefix, suffix string) *authtypes.ModuleAccount {
	header := ctx.BlockHeader()

	buf := []byte(types.ModuleName + prefix)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	addrBuf := string(buf) + suffix

	return authtypes.NewEmptyModuleAccount(addrBuf, authtypes.Staking)
}

func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []types.DAPair) {
	store := ctx.KVStore(k.storeKey)

	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, currTime)
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := types.DAPairs{}
		value := unbondingTimesliceIterator.Value()
		k.cdc.MustUnmarshal(value, &timeslice)

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)

		store.Delete(unbondingTimesliceIterator.Key())
	}

	return matureUnbonds
}

func (k Keeper) ProcessCompletedUnbonding(ctx sdk.Context) {
	matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
	for _, dvPair := range matureUnbonds {
		_, err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddress, dvPair.AgentId)
		if err != nil {
			continue
		}
	}
}

func (k Keeper) CompleteUnbonding(ctx sdk.Context, delegator string, agentID uint64) (sdk.Coins, error) {
	ubd, found := k.GetMultiStakingUnbonding(ctx, agentID, delegator)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	balances := sdk.NewCoins()
	ctxTime := ctx.BlockHeader().Time

	delegatorAddress, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) {
			ubd.RemoveEntry(int64(i))
			i--

			if !entry.Balance.IsZero() {

				k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAddress, sdk.Coins{entry.Balance})
				balances = balances.Add(entry.Balance)
			}
		}
	}

	if len(ubd.Entries) == 0 {
		k.RemoveMultiStakingUnbonding(ctx, agentID, delegator)
	} else {
		k.SetMultiStakingUnbonding(ctx, agentID, delegator, ubd)
	}

	return balances, nil
}

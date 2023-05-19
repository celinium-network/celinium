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
	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.DelegateAddress)

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

	return k.mintAndDelegate(ctx, agent, *validator, bondTokenAmt)
}

func (k Keeper) mintAndDelegate(ctx sdk.Context, agent *types.MultiStakingAgent, validator stakingtypes.Validator, amount sdk.Coin) error {
	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.DelegateAddress)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{amount}); err != nil {
		return err
	}

	k.bankKeeper.AddSupplyOffset(ctx, amount.Denom, amount.Amount)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, agentDelegateAccAddr, sdk.Coins{amount}); err != nil {
		return err
	}

	if _, err := k.stakingkeeper.Delegate(ctx,
		agentDelegateAccAddr, amount.Amount,
		stakingtypes.Unbonded, validator, true,
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
	removeShares := agent.CalculateShares(msg.Amount.Amount)
	if err := k.DecreaseMultiStakingShares(ctx, removeShares, agent.Id, msg.DelegatorAddress); err != nil {
		return err
	}

	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)
	undelegateAmt, err := k.EquivalentCoinCalculator(ctx, msg.Amount, defaultBondDenom)
	if err != nil {
		return err
	}

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return err
	}

	if err := k.undelegateAndBurn(ctx, agent, valAddr, undelegateAmt); err != nil {
		return err
	}

	unbonding := k.GetOrCreateMultiStakingUnbonding(ctx, agent.Id, msg.DelegatorAddress)
	unbondingTime := k.stakingkeeper.GetParams(ctx).UnbondingTime

	// TODO Whether the length of Entry should be limited ?
	undelegateCompleteTime := ctx.BlockTime().Add(unbondingTime)
	unbonding.Entries = append(unbonding.Entries, types.MultiStakingUnbondingEntry{
		CompletionTime: undelegateCompleteTime,
		InitialBalance: msg.Amount,
		Balance:        msg.Amount,
	})

	k.SetMultiStakingUnbonding(ctx, agent.Id, msg.DelegatorAddress, unbonding)

	agent.Shares = agent.Shares.Sub(removeShares)
	agent.StakedAmount = agent.StakedAmount.Sub(msg.Amount.Amount)

	k.SetMultiStakingAgent(ctx, agent)
	k.InsertUBDQueue(ctx, unbonding, undelegateCompleteTime)

	return nil
}

func (k Keeper) undelegateAndBurn(ctx sdk.Context, agent *types.MultiStakingAgent, valAddr sdk.ValAddress, undelegateAmt sdk.Coin) error {
	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.DelegateAddress)

	stakedShares, err := k.stakingkeeper.ValidateUnbondAmount(ctx, agentDelegateAccAddr, valAddr, undelegateAmt.Amount)
	if err != nil {
		return err
	}

	undelegationCoins, err := k.stakingkeeper.InstantUndelegate(ctx, agentDelegateAccAddr, valAddr, stakedShares)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		agentDelegateAccAddr, types.ModuleName, undelegationCoins,
	); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, undelegationCoins); err != nil {
		return err
	}
	k.bankKeeper.AddSupplyOffset(ctx, undelegationCoins[0].Denom, undelegationCoins[0].Amount)

	return nil
}

func (k Keeper) agentValidator(ctx sdk.Context, agent *types.MultiStakingAgent) (*stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(agent.ValidatorAddress)
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
		Id:               newAgentID,
		StakeDenom:       denom,
		DelegateAddress:  newAccount.Address,
		ValidatorAddress: valAddr,
		WithdrawAddress:  newAccount.Address,
		StakedAmount:     math.ZeroInt(),
		RewardAmount:     math.ZeroInt(),
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

	agent, found := k.GetMultiStakingAgentByID(ctx, agentID)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	agentDelegateAddress := sdk.MustAccAddressFromBech32(agent.DelegateAddress)

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
				k.sendCoinsFromAccountToAccount(ctx, agentDelegateAddress, delegatorAddress, sdk.Coins{entry.Balance})
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

func (k Keeper) RefreshAgentDelegationAmount(ctx sdk.Context) {
	agents := k.GetAllAgent(ctx)

	for i := 0; i < len(agents); i++ {
		valAddress, err := sdk.ValAddressFromBech32(agents[i].ValidatorAddress)
		if err != nil {
			panic(err)
		}

		validator, found := k.stakingkeeper.GetValidator(ctx, valAddress)
		if !found {
			continue
		}

		var currentAmount math.Int
		delegator := sdk.MustAccAddressFromBech32(agents[i].DelegateAddress)
		delegation, found := k.stakingkeeper.GetDelegation(ctx, delegator, valAddress)
		if !found {
			continue
		} else {
			currentAmount = validator.TokensFromShares(delegation.Shares).RoundInt()
		}
		refreshedAmount, _ := k.GetExpectedDelegationAmount(ctx, sdk.NewCoin(agents[i].StakeDenom, agents[i].StakedAmount))

		if refreshedAmount.Amount.GT(currentAmount) {
			adjustment := refreshedAmount.Amount.Sub(currentAmount)
			k.mintAndDelegate(ctx, &agents[i], validator, sdk.NewCoin(refreshedAmount.Denom, adjustment))
		} else if refreshedAmount.Amount.LT(currentAmount) {
			adjustment := currentAmount.Sub(refreshedAmount.Amount)
			k.undelegateAndBurn(ctx, &agents[i], valAddress, sdk.NewCoin(refreshedAmount.Denom, adjustment))
		}
	}
}

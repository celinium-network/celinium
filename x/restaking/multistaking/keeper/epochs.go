package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/celinium-network/celinium/x/epochs/types"
	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, _ string, _ int64) {
}

// BeforeEpochStart implements types.EpochHooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	switch epochIdentifier {
	case types.RefreshAgentDelegationEpochID:
		h.k.RefreshAgentDelegationAmount(ctx)
	case types.CollectAgentStakingRewardEpochID:
		// TODO remove it from epoch ?
		h.k.CollectAgentsReward(ctx)
	}
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

func (k Keeper) CollectAgentsReward(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.MultiStakingAgentPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var agent types.MultiStakingAgent
		// TODO panic or continue ?
		err := k.cdc.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}

		delegator := sdk.MustAccAddressFromBech32(agent.DelegateAddress)
		valAddr, err := sdk.ValAddressFromBech32(agent.ValidatorAddress)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("convert validator address from bech32:%s failed, err: %s", agent.ValidatorAddress, err))
			continue
		}
		rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, delegator, valAddr)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Withdraw delegation reward failed. AgentID: %d", agent.Id))
			continue
		}

		// TODO multi kind reward coins
		agent.RewardAmount = agent.RewardAmount.Add(rewards[0].Amount)
		agentBz, err := k.cdc.Marshal(&agent)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}
		store.Set(iterator.Key(), agentBz)
	}
}

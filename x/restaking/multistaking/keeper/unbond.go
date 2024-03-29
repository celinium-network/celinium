package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/x/restaking/multistaking/types"
)

func (k Keeper) ProcessCompletedUnbonding(ctx sdk.Context) {
	matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
	for _, dvPair := range matureUnbonds {
		_, err := k.CompleteUnbonding(ctx, dvPair.DelegatorAddress, dvPair.AgentId)
		if err != nil {
			continue
		}
	}
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

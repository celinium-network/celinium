package keeper

import (
	"strings"

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

	undelegationCoins, err := k.instantUndelegate(ctx, agentDelegateAccAddr, valAddr, stakedShares)
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
		Shares:           math.ZeroInt(),
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

func (k Keeper) instantUndelegate(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
) (sdk.Coins, error) {
	validator, found := k.stakingkeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingkeeper.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return nil, err
	}

	bondDenom := k.stakingkeeper.GetParams(ctx).BondDenom

	amt := sdk.NewCoin(bondDenom, returnAmount)
	res := sdk.NewCoins(amt)

	moduleName := stakingtypes.NotBondedPoolName
	if validator.IsBonded() {
		moduleName = stakingtypes.BondedPoolName
	}
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, moduleName, delAddr, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

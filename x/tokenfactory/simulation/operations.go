package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"celinium/x/tokenfactory/keeper"
	"celinium/x/tokenfactory/types"

	simhepler "celinium/app/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateDenom = "op_weight_msg_create_denom" //nolint:gosec
	OpWeightMsgMintDenom   = "op_weight_msg_mint_denom"   //nolint:gosec
	OpWeightMsgBurnDenom   = "op_weight_msg_burn_denom"   //nolint:gosec
	OpWeightMsgChangeAdmin = "op_weight_msg_change_admin" //nolint:gosec
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgCreateDenom, weightMsgMintDenom, weightMsgBurnDenom, weightMsgChangeAdmin int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateDenom, &weightMsgCreateDenom, nil,
		func(_ *rand.Rand) {
			weightMsgCreateDenom = simappparams.DefaultWeightMsgSend
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgMintDenom, &weightMsgMintDenom, nil,
		func(_ *rand.Rand) {
			weightMsgMintDenom = simappparams.DefaultWeightMsgSend
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgBurnDenom, &weightMsgBurnDenom, nil,
		func(_ *rand.Rand) {
			weightMsgBurnDenom = simappparams.DefaultWeightMsgSend
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgChangeAdmin, &weightMsgChangeAdmin, nil,
		func(_ *rand.Rand) {
			weightMsgChangeAdmin = simappparams.DefaultWeightMsgSend
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateDenom,
			SimulateMsgCreateDenom(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgMintDenom,
			SimulateMsgMintDenom(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBurnDenom,
			SimulateMsgBurnDenom(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgChangeAdmin,
			SimulateMsgChangeAdmin(ak, bk, k),
		),
	}
}

func SimulateMsgCreateDenom(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			fees sdk.Coins
			err  error
		)

		minCoins := k.GetParams(ctx).DenomCreationFee

		creator, _ := simhepler.RandomSimAccountWithMinCoins(bk, ctx, r, accs, minCoins)
		msg := &types.MsgCreateDenom{
			Sender:   creator.Address.String(),
			Subdenom: simhepler.RandStringOfLength(r, types.MaxSubdenomLength),
		}

		spendable := bk.SpendableCoins(ctx, creator.Address)
		coins, hasNeg := spendable.SafeSub(minCoins...)
		if hasNeg {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateDenom, "no enough mint fee"), nil, err
		}

		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateDenom, "no enough fees"), nil, err
		}

		err = simhepler.SendMsgSend(r, app, bk, k, ak, msg, ctx, chainID, creator.Address, []cryptotypes.PrivKey{creator.PrivKey}, fees)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid create denom"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgMintDenom(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			err error
		)

		acc, senderExists := simhepler.RandomSimAccountWithConstraint(r, accountCreatedTokenFactoryDenom(k, ctx), accs)
		if !senderExists {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMint, "no creator"), nil, err
		}

		denom, addr, err := getTokenFactoryDenomAndItsAdmin(r, k, ctx, acc)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMint, "get admin factory token failed"), nil, err
		}

		if addr == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMint, "denom has no admin"), nil, err
		}

		var adminAccount simtypes.Account
		for _, c := range accs {
			if c.Address.Equals(addr) {
				adminAccount = c
			}
		}

		mintAmount := simhepler.RandPositiveInt(r, sdk.NewIntFromUint64(1000_000000))

		msg := &types.MsgMint{
			Sender: adminAccount.Address.String(),
			Amount: sdk.NewCoin(denom, mintAmount),
		}

		fees := simhepler.RandomFees(r, ctx, bk, adminAccount.Address, sdk.DefaultBondDenom)

		err = simhepler.SendMsgSend(r, app, bk, k, ak, msg, ctx, chainID, adminAccount.Address, []cryptotypes.PrivKey{adminAccount.PrivKey}, fees)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid mint denom"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgBurnDenom(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			fees sdk.Coins
			err  error
		)

		acc, senderExists := simhepler.RandomSimAccountWithConstraint(r, accountCreatedTokenFactoryDenom(k, ctx), accs)
		if !senderExists {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBurn, "no creator"), nil, err
		}

		denom, addr, err := getTokenFactoryDenomAndItsAdmin(r, k, ctx, acc)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBurn, "get admin factory token failed"), nil, err
		}

		if addr == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBurn, "denom has no admin"), nil, err
		}

		denomBal := simhepler.GetSpendableBalance(ctx, bk, addr, denom)
		if denomBal == nil || denomBal.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBurn, "addr does not have enough balance to burn"), nil, err
		}

		var adminAccount simtypes.Account
		for _, c := range accs {
			if c.Address.Equals(addr) {
				adminAccount = c
			}
		}

		burnAmount := simhepler.RandPositiveInt(r, denomBal.Amount)
		msg := &types.MsgBurn{
			Sender: adminAccount.Address.String(),
			Amount: sdk.NewCoin(denom, burnAmount),
		}

		fees = simhepler.RandomFees(r, ctx, bk, adminAccount.Address, sdk.DefaultBondDenom)

		err = simhepler.SendMsgSend(r, app, bk, k, ak, msg, ctx, chainID, adminAccount.Address, []cryptotypes.PrivKey{adminAccount.PrivKey}, fees)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid burn denom"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgChangeAdmin(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			fees sdk.Coins
			err  error
		)

		acc, senderExists := simhepler.RandomSimAccountWithConstraint(r, accountCreatedTokenFactoryDenom(k, ctx), accs)
		if !senderExists {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeAdmin, "no creator"), nil, err
		}

		denom, addr, err := getTokenFactoryDenomAndItsAdmin(r, k, ctx, acc)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeAdmin, "get admin factory token failed"), nil, err
		}

		if addr == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeAdmin, "denom has no admin"), nil, err
		}

		minCoins := k.GetParams(ctx).DenomCreationFee
		newAdmin, _ := simhepler.RandomSimAccountWithMinCoins(bk, ctx, r, accs, minCoins)

		fees = simhepler.RandomFees(r, ctx, bk, addr, sdk.DefaultBondDenom)

		msg := &types.MsgChangeAdmin{
			Sender:   addr.String(),
			Denom:    denom,
			NewAdmin: newAdmin.Address.String(),
		}

		var newAcc simtypes.Account
		for _, c := range accs {
			if c.Address.Equals(addr) {
				newAcc = c
			}
		}

		err = simhepler.SendMsgSend(r, app, bk, k, ak, msg, ctx, chainID, addr, []cryptotypes.PrivKey{newAcc.PrivKey}, fees)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid change admin"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func accountCreatedTokenFactoryDenom(k keeper.Keeper, ctx sdk.Context) simhepler.SimAccountConstraint {
	return func(acc simtypes.Account) bool {
		store := k.GetCreatorPrefixStore(ctx, acc.Address.String())
		iterator := store.Iterator(nil, nil)
		defer iterator.Close()
		return iterator.Valid()
	}
}

func getTokenFactoryDenomAndItsAdmin(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, acc simtypes.Account) (string, sdk.AccAddress, error) {
	store := k.GetCreatorPrefixStore(ctx, acc.Address.String())
	denoms := simhepler.GatherAllKeysFromStore(store)

	denom := simhepler.RandSelect(r, denoms...)

	authData, err := k.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return "", nil, err
	}
	admin := authData.Admin
	addr, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return "", nil, err
	}
	return denom, addr, nil
}

package simulation

import (
	"errors"
	"math/big"
	"math/rand"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"celinium/x/tokenfactory/keeper"
	"celinium/x/tokenfactory/types"
	tftypes "celinium/x/tokenfactory/types"
)

func GatherAllKeysFromStore(storeObj store.KVStore) []string {
	iterator := storeObj.Iterator(nil, nil)
	defer iterator.Close()

	keys := []string{}
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, string(iterator.Key()))
	}
	return keys
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringOfLength generates a random string of a particular length
func RandStringOfLength(r *rand.Rand, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func RandSelect[T interface{}](r *rand.Rand, args ...T) T {
	idx := r.Intn(len(args))
	return args[idx]
}

func RandPositiveInt(r *rand.Rand, max sdk.Int) sdk.Int {
	if !max.GTE(sdk.OneInt()) {
		panic("max too small")
	}

	max = max.Sub(sdk.OneInt())

	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt())
}

func GetSpendableBalance(ctx sdk.Context, bk tftypes.BankKeeper, sender sdk.AccAddress, denom string) *sdk.Coin {
	spendable := bk.SpendableCoins(ctx, sender)
	for _, c := range spendable {
		if c.Denom == denom {
			return &c
		}
	}
	return nil
}

func RandomFees(r *rand.Rand, ctx sdk.Context, bk tftypes.BankKeeper, sender sdk.AccAddress, denom string) []sdk.Coin {
	spendableBalance := GetSpendableBalance(ctx, bk, sender, sdk.DefaultBondDenom)
	if spendableBalance == nil || spendableBalance.Amount.IsZero() {
		return nil
	}

	fees, err := simtypes.RandomFees(r, ctx, sdk.Coins{*spendableBalance})
	if err != nil {
		return nil
	}
	return fees
}

func RandomSimAccountWithMinCoins(bk types.BankKeeper, ctx sdk.Context, r *rand.Rand, accs []simtypes.Account, coins sdk.Coins) (simtypes.Account, error) {
	accHasMinCoins := func(acc simtypes.Account) bool {
		for _, c := range coins {
			if !bk.HasBalance(ctx, acc.Address, c) {
				return false
			}
		}
		return true
	}
	acc, found := RandomSimAccountWithConstraint(r, accHasMinCoins, accs)
	if !found {
		return simtypes.Account{}, errors.New("no address with min balance found")
	}
	return acc, nil
}

type SimAccountConstraint = func(account simtypes.Account) bool

// returns acc, accExists := sim.RandomSimAccountWithConstraint(f)
// where acc is a uniformly sampled account from all accounts satisfying the constraint f
// a constraint is satisfied for an account `acc` if f(acc) = true
// accExists is false, if there is no such account.
func RandomSimAccountWithConstraint(r *rand.Rand, f SimAccountConstraint, accs []simtypes.Account) (simtypes.Account, bool) {
	filteredAddrs := []simtypes.Account{}
	for _, acc := range accs {
		if f(acc) {
			filteredAddrs = append(filteredAddrs, acc)
		}
	}

	if len(filteredAddrs) == 0 {
		return simtypes.Account{}, false
	}

	idx := r.Intn(len(filteredAddrs))
	return filteredAddrs[idx], true
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func SendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp, bk tftypes.BankKeeper, k keeper.Keeper, ak tftypes.AccountKeeper,
	msg sdk.Msg, ctx sdk.Context, chainID string, sender sdk.AccAddress, privkeys []cryptotypes.PrivKey, fees sdk.Coins,
) error {
	var (
		err error
	)

	account := ak.GetAccount(ctx, sender)

	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return err
	}

	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}

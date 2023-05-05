package types_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func TestSourceChainAllocateFunds(t *testing.T) {
	srcChains := []types.SourceChain{
		{
			Validators: []types.Validator{
				{
					"validator1",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
			},
		},
		{
			Validators: []types.Validator{
				{
					"validator1",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
				{
					"validator2",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
			},
		},
		{
			Validators: []types.Validator{
				{
					"validator1",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
				{
					"validator2",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
				{
					"validator3",
					sdk.ZeroInt(),
					rand.Uint64()%100000 + types.MinValidatorWeight, //nolint:gosec
				},
			},
		},
	}

	totalFunds := sdk.NewIntFromUint64(rand.Uint64()%100000 + 1010101001001) //nolint:gosec

	checkAlloc := func(srcChain *types.SourceChain, funds math.Int, allocVals types.Validators) {
		if len(allocVals.Validators) != len(srcChain.Validators) {
			t.Fatal("allocate fund error")
		}
		totalAmount := sdk.ZeroInt()
		for _, f := range allocVals.Validators {
			totalAmount = totalAmount.Add(f.TokenAmount)
		}
		if !totalAmount.Equal(funds) {
			t.Fatal("allocate amount not equal")
		}
	}

	for i := 0; i < len(srcChains); i++ {
		allocFunds := srcChains[i].AllocateTokenForValidator(totalFunds)
		checkAlloc(&srcChains[i], totalFunds, allocFunds)
	}
}

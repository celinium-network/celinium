package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	MinValidatorWeight = uint64(1000)
	MinValidators      = 1

	// WithdrawAddressSuffix used for generate liquidstake withdraw address
	WithdrawAddressSuffix = "withdraw"
	// DelegationAddressSuffix used for generate liquidstake delegateion address
	DelegationAddressSuffix = "delegate"
	// UnboundingAddressSuffix used for generate liquidstake unbond address
	UnboundAddressSuffix = "unbounding"
)

// BasicVerify verify SouceChain parameters
// todo: more verify ?
func (s SourceChain) BasicVerify() error {
	if len(s.Validators) < MinValidators {
		return fmt.Errorf("min validators: %d, get: %d", MinValidators, len(s.Validators))
	}

	for _, v := range s.Validators {
		if !verifyValidatorAddress(v.Address, s.Bech32ValidatorAddrPrefix) {
			return fmt.Errorf("invalid validator address of souce chain, Address: %s", v.Address)
		}

		if v.Weight <= MinValidatorWeight {
			return fmt.Errorf("min weight: %d, get: %d from validators: %s", MinValidatorWeight, v.Weight, v.Address)
		}
	}

	if s.NativeDenom == s.DerivativeDenom {
		return fmt.Errorf("NativeDenom equal DerivativeDenom")
	}

	if err := sdk.ValidateDenom(s.NativeDenom); err != nil {
		return err
	}

	return sdk.ValidateDenom(s.DerivativeDenom)
}

// GenerateAccounts generate the WithdrawAddress/DelegateAddress/UnboudAddress for source chain
// TODO: Add a function parameter to do some work like register, then don't need return values?
func (s *SourceChain) GenerateAccounts(ctx sdk.Context) (accounts []*authtypes.ModuleAccount) {
	header := ctx.BlockHeader()

	buf := []byte(ModuleName + s.ChainID + s.ConnectionID)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	withdrawAddrBuf := string(buf) + WithdrawAddressSuffix
	withDrawAccount := authtypes.NewEmptyModuleAccount(withdrawAddrBuf, authtypes.Staking)

	delegationAddrBuf := string(buf) + DelegationAddressSuffix
	ecsrowAccount := authtypes.NewEmptyModuleAccount(delegationAddrBuf, authtypes.Staking)

	unbondAddrBuf := string(buf) + UnboundAddressSuffix
	unbondAccount := authtypes.NewEmptyModuleAccount(unbondAddrBuf, authtypes.Staking)

	s.WithdrawAddress = withDrawAccount.Address
	s.EcsrowAddress = ecsrowAccount.Address
	s.DelegateAddress = unbondAccount.Address

	accounts = append(accounts, withDrawAccount)
	accounts = append(accounts, unbondAccount)

	return accounts
}

func (s SourceChain) AllocateFundsForValidator(amount math.Int) map[string]math.Int {
	validatorFunds := make(map[string]math.Int)

	// TODO weight shoudle math.Int, maybe overflow there?
	var totalWeight uint64
	for _, v := range s.Validators {
		totalWeight += v.Weight
	}

	// TODO the last validator get all remind funds
	for _, v := range s.Validators {
		allocateFundAmount := amount.Mul(math.NewIntFromUint64(v.Weight)).Quo(math.NewIntFromUint64(totalWeight))
		validatorFunds[v.Address] = allocateFundAmount
	}

	return validatorFunds
}

func (s *SourceChain) UpdateWithDelegationRecord(record *DelegationRecord) {
	s.StakedAmount = s.StakedAmount.Add(record.DelegationCoin.Amount)
	// TODO update delegation amout for every validators, it't will be used for rebalance.
	// (1) should not calcaute from weight at now
	// (2) record at callback'Args
	// (3) only after successful delegation
}

func verifyValidatorAddress(address, addrPrefix string) bool {
	bz, err := sdk.GetFromBech32(address, addrPrefix)
	if err != nil {
		return false
	}

	err = sdk.VerifyAddressFormat(bz)
	return err == nil
}

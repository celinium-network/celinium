package types

import (
	"fmt"

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
	if len(s.Validators) <= MinValidators {
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

// GenerateAndFillAccount generate the WithdrawAddress/DelegateAddress/UnboudAddress for source chain
// todo!: Add a function parameter to do some work like register, then don't need return values?
func (s *SourceChain) GenerateAndFillAccount(ctx sdk.Context) (accounts []*authtypes.ModuleAccount) {
	header := ctx.BlockHeader()

	buf := []byte(ModuleName + s.ChainID + s.ConnectionID)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	withdrawAddrBuf := string(buf) + WithdrawAddressSuffix
	withDrawAccount := authtypes.NewEmptyModuleAccount(withdrawAddrBuf, authtypes.Staking)

	delegationAddrBuf := string(buf) + DelegationAddressSuffix
	delegationAccount := authtypes.NewEmptyModuleAccount(delegationAddrBuf, authtypes.Staking)

	unbondAddrBuf := string(buf) + UnboundAddressSuffix
	unbondAccount := authtypes.NewEmptyModuleAccount(unbondAddrBuf, authtypes.Staking)

	s.WithdrawAddress = withDrawAccount.Address
	s.DelegateAddress = delegationAccount.Address
	s.UnboudAddress = unbondAccount.Address

	accounts = append(accounts, withDrawAccount)
	accounts = append(accounts, delegationAccount)
	accounts = append(accounts, unbondAccount)

	return accounts
}

func verifyValidatorAddress(address, addrPrefix string) bool {
	bz, err := sdk.GetFromBech32(address, addrPrefix)
	if err != nil {
		return false
	}

	err = sdk.VerifyAddressFormat(bz)
	return err == nil
}

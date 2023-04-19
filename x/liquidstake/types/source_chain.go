package types

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	transfertype "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

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

func (s *SourceChain) GenerateIBCDeonm() error {
	parts := []string{transfertype.PortID, s.TransferChannelID, s.NativeDenom}
	denom := strings.Join(parts, "/")
	denomTrace := transfertype.ParseDenomTrace(denom)
	if err := denomTrace.Validate(); err != nil {
		return err
	}

	s.IbcDenom = denomTrace.IBCDenom()

	return nil
}

type ValidatorFund struct {
	Address string
	Amount  math.Int
}

func (s SourceChain) AllocateFundsForValidator(amount math.Int) []ValidatorFund {
	validatorFunds := make([]ValidatorFund, 0)

	totalWeight := math.ZeroInt()
	for _, v := range s.Validators {
		totalWeight = totalWeight.Add(math.NewIntFromUint64(v.Weight))
	}

	valiLen := len(s.Validators)
	reminding := amount
	for i := 0; i < valiLen-1; i++ {
		allocateAmt := amount.Mul(math.NewIntFromUint64(s.Validators[i].Weight)).Quo(totalWeight)
		validatorFunds = append(validatorFunds, ValidatorFund{
			Address: s.Validators[i].Address,
			Amount:  allocateAmt,
		})
		reminding = reminding.Sub(allocateAmt)
	}

	// the last validator get all reminding amount
	validatorFunds = append(validatorFunds, ValidatorFund{
		Address: s.Validators[valiLen-1].Address,
		Amount:  reminding,
	})

	return validatorFunds
}

func (s *SourceChain) UpdateWithDelegationRecord(record *DelegationRecord) {
	s.StakedAmount = s.StakedAmount.Add(record.DelegationCoin.Amount)
	// TODO update delegation amout for every validators, it't will be used for rebalance.
	// (1) should not calcaute from weight at now
	// (2) record at callback'Args
	// (3) only after successful delegation
}

func (s SourceChain) ValidatorsAddress() string {
	var vs []string
	for _, v := range s.Validators {
		vs = append(vs, v.Address)
	}
	return strings.Join(vs, ",")
}

func (s SourceChain) ValidatorsWeight() string {
	var vw []string
	for _, v := range s.Validators {
		vw = append(vw, strconv.FormatUint(v.Weight, 10))
	}
	return strings.Join(vw, ",")
}

func verifyValidatorAddress(address, addrPrefix string) bool {
	bz, err := sdk.GetFromBech32(address, addrPrefix)
	if err != nil {
		return false
	}

	err = sdk.VerifyAddressFormat(bz)
	return err == nil
}

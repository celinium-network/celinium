package types

// DelegationRecord status
const (
	DelegationPending = iota
	DelegationTransferring
	DelegationTransferred
	Delegating
	DelegationDone
	DelegationTransferFailed
	DelegateFailed
)

// UndelegationRecord status
const (
	UndelegationPending = iota
	UndelegationClaimable
	UndelegationComplete
)

// Unbonding failed
const (
	UnbondingPending = iota
	UnbondingStart
	UnbondingWaitting
	UnbondingWithdraw
	UnbondingTransferred
	UnbondingDone
	UnbondingStartFailed
	UnbondingTransferFailed
)

package types

// TODO: More precise and less confusing enums.

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

// TODO where the funds when status is DelegationTransferFailed and DelegateFailed
// (1) When status equal DelegationTransferFailed, the funds will be refund to the ibc msg sender.
// Maybe we don't need the status of DelegationTransferFailed, just set it to DelegationPending, and start over from scratch.
// (2) When status equal DelegateFailed, the funds will is locked in source chain.
func IsDelegationRecordProcessing(status int) bool {
	return status != DelegationDone
}

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

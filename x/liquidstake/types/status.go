package types

// DelegationRecord status
type DelegationRecordStatus uint32

const (
	DelegationPending DelegationRecordStatus = iota
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
func IsDelegationRecordProcessing(status DelegationRecordStatus) bool {
	return status != DelegationDone
}

// UndelegationRecord status
type UndelegationRecordStatus uint32

const (
	UndelegationPending UndelegationRecordStatus = iota
	UndelegationClaimable
	UndelegationComplete
)

// Unbonding failed
type UnbondingStatus uint32

const (
	UnbondingPending UnbondingStatus = iota
	UnbondingStart
	UnbondingWaitting
	UnbondingWithdraw
	UnbondingTransferred
	UnbondingDone
	UnbondingStartFailed
	UnbondingTransferFailed
)

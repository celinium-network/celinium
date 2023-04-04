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

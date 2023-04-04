package types

// DelegationRecord status
const (
	Pending = iota
	Transferring
	Transferred
	Delegating
	Done
	TransferFailed
	DelegateFailed
)

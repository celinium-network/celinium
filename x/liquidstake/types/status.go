package types

// ProxyDelegation status
type ProxyDelegationStatus uint32

const (
	ProxyDelegationPending ProxyDelegationStatus = iota
	ProxyDelegationTransferring
	ProxyDelegationTransferred
	ProxyDelegating
	ProxyDelegationDone
	ProxyDelegationTransferFailed
	ProxyDelegationFailed
)

// TODO where the funds when status is DelegationTransferFailed and DelegateFailed
// (1) When status equal DelegationTransferFailed, the funds will be refund to the ibc msg sender.
// Maybe we don't need the status of DelegationTransferFailed, just set it to DelegationPending, and start over from scratch.
// (2) When status equal DelegateFailed, the funds will is locked in source chain.
func IsProxyDelegationProcessing(status ProxyDelegationStatus) bool {
	return status != ProxyDelegationDone
}

// UndelegationRecord status
type UserUnbondingStatus uint32

const (
	UserUnbondingPending UserUnbondingStatus = iota
	UserUnbondingClaimable
	UserUnbondingComplete
)

// Unbonding failed
type ProxyUnbondingStatus uint32

const (
	ProxyUnbondingPending ProxyUnbondingStatus = iota
	ProxyUnbondingStart
	ProxyUnbondingWaitting
	ProxyUnbondingWithdraw
	ProxyUnbondingTransferred
	ProxyUnbondingDone
	ProxyUnbondingStartFailed
	ProxyUnbondingTransferFailed
)

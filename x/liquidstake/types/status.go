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

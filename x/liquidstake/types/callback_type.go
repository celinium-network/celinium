package types

type CallType uint32

// Callback type
const (
	DelegateTransferCall CallType = iota
	DelegateCall
	UnbondCall
	WithdrawUnbondCall
	WithdrawDelegateRewardCall
	TransferRewardCall
	SetWithdrawAddressCall
)

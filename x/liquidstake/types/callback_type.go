package types

type CallType uint32

// Callback type
const (
	DelegateTransferCall CallType = iota
	DelegateCall
	UndelegateCall
	WithdrawUnbondCall
	WithdrawDelegateRewardCall
	TransferRewardCall
	SetWithdrawAddressCall
)

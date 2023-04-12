package types

import (
	sdkioerrors "cosmossdk.io/errors"
)

var (
	ErrSourceChainExist        = sdkioerrors.Register(ModuleName, 1, "source chain already exist")
	ErrBannedIBCTransfer       = sdkioerrors.Register(ModuleName, 2, "ibc transfer is banned")
	ErrSourceChainParameter    = sdkioerrors.Register(ModuleName, 3, "add source chain parameter has error")
	ErrUnknownEpoch            = sdkioerrors.Register(ModuleName, 4, "unknown epoch")
	ErrNoExistDelegationRecord = sdkioerrors.Register(ModuleName, 5, "delegation record don't exist")
	ErrUnknownSourceChain      = sdkioerrors.Register(ModuleName, 6, "source chain is not exist")
	ErrRepeatUndelegate        = sdkioerrors.Register(ModuleName, 7, "repeattly delegate in a epoch")
	ErrInternalError           = sdkioerrors.Register(ModuleName, 8, "internal error")
	ErrInsufficientFunds       = sdkioerrors.Register(ModuleName, 9, "insufficient funds to support the current transaction")
	ErrEpochUnbondingNotExist  = sdkioerrors.Register(ModuleName, 10, "unbondings not found for specific epoch")
	ErrCallbackNotExist        = sdkioerrors.Register(ModuleName, 11, "callback not find")
	ErrIBCQueryNotExist        = sdkioerrors.Register(ModuleName, 12, "ibcquery not exist")
)

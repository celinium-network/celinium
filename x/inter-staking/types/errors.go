package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrSourceChainExist       = sdkerrors.Register(ModuleName, 2, "source chain already exist")
	ErrUnknownSourceChain     = sdkerrors.Register(ModuleName, 3, "unknown source chain")
	ErrMismatchParameter      = sdkerrors.Register(ModuleName, 4, "parameters in msg has error")
	ErrUnavailableSourceChain = sdkerrors.Register(ModuleName, 5, "unavailable source chain")
	ErrMismatchSourceCoin     = sdkerrors.Register(ModuleName, 6, "mismatch source chain coin")
	ErrInsufficientDelegation = sdkerrors.Register(ModuleName, 7, "insufficient delegation amount")
)

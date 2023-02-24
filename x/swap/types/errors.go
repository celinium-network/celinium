package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidTokens            = sdkerrors.Register(ModuleName, 1, "invalid token")
	ErrPairNotExist             = sdkerrors.Register(ModuleName, 2, "pair not exist")
	ErrPairCreated              = sdkerrors.Register(ModuleName, 3, "pair already created")
	ErrInvalidTokenAmountRange  = sdkerrors.Register(ModuleName, 4, "invalid token amount range")
	ErrInsufficientFunds        = sdkerrors.Register(ModuleName, 5, "insufficient token amount")
	ErrMath                     = sdkerrors.Register(ModuleName, 6, "error occor in calculate")
	ErrInvalidPath              = sdkerrors.Register(ModuleName, 7, "Invalid swap path")
	ErrInsufficientTargetAmount = sdkerrors.Register(ModuleName, 8, "insufficient target amount after swap")
	ErrMismatchParameter        = sdkerrors.Register(ModuleName, 9, "mistach parameter")
	ErrInsufficientPairReserve  = sdkerrors.Register(ModuleName, 10, "pair liquidity is insufficient")
)

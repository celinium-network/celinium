package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreatePair               = "create_pair"
	TypeMsgAddLiquidty              = "add_liquidty"
	TypeMsgSwapExactTokensForTokens = "swap_exact_tokens_for_tokens"
)

var _ sdk.Msg = &MsgCreatePair{}

func NewMsgCreatePair(sender, token0, token1 string) *MsgCreatePair {
	return &MsgCreatePair{
		Sender: sender,
		Token0: token0,
		Token1: token1,
	}
}

func (m MsgCreatePair) Route() string { return RouterKey }

func (m MsgCreatePair) Type() string { return TypeMsgCreatePair }

func (m MsgCreatePair) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if len(m.Token0) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "token0:(%s) is nil", m.Token0)
	}

	if m.Token0 == m.Token1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "token0 == token1")
	}

	return nil
}

func (m MsgCreatePair) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCreatePair) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgAddLiquidity{}

func (m MsgAddLiquidity) Route() string { return RouterKey }

func (m MsgAddLiquidity) Type() string { return TypeMsgAddLiquidty }

func (m MsgAddLiquidity) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if err := sdk.ValidateDenom(m.Token0.Denom); err != nil {
		return err
	}

	if err := sdk.ValidateDenom(m.Token1.Denom); err != nil {
		return err
	}

	return nil
}

func (m MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

var _ sdk.Msg = &MsgSwapExactTokensForTokens{}

func (m MsgSwapExactTokensForTokens) Route() string { return RouterKey }

func (m MsgSwapExactTokensForTokens) Type() string { return TypeMsgSwapExactTokensForTokens }

func (m MsgSwapExactTokensForTokens) ValidateBasic() error {

	return nil
}

func (m MsgSwapExactTokensForTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgSwapExactTokensForTokens) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}

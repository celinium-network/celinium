package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgRegisterSourceChain{}
	_ sdk.Msg = &MsgEditVadlidators{}
	_ sdk.Msg = &MsgRebalanceValidators{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUndelegate{}
	_ sdk.Msg = &MsgReinvest{}
	_ sdk.Msg = &MsgClaim{}
)

// GetSigners implements types.Msg
func (msg *MsgRegisterSourceChain) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgRegisterSourceChain) ValidateBasic() error {
	return nil
}

func (msg *MsgEditVadlidators) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgEditVadlidators) ValidateBasic() error {
	return nil
}

func (msg *MsgRebalanceValidators) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgRebalanceValidators) ValidateBasic() error {
	return nil
}

func (msg *MsgDelegate) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgDelegate) ValidateBasic() error {
	return nil
}

func (msg *MsgUndelegate) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgUndelegate) ValidateBasic() error {
	return nil
}

func (msg *MsgReinvest) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Caller)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgReinvest) ValidateBasic() error {
	return nil
}

func (msg *MsgClaim) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic implements types.Msg
func (*MsgClaim) ValidateBasic() error {
	return nil
}

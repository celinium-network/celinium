package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgAddSourceChain{}
	_ sdk.Msg = &MsgUpdateSourceChainDelegatePlan{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUnDelegate{}
)

func NewMsgAddSourceChain(
	chainID string,
	connectionID string,
	version string,
	sourceChainDenom string,
	sourceChainTraceDenom string,
	strategy []DelegationStrategy,
	authority string,
) *MsgAddSourceChain {
	return &MsgAddSourceChain{
		ChainId:               chainID,
		ConnectionId:          connectionID,
		Version:               version,
		SourceChainDenom:      sourceChainDenom,
		SourceChainTraceDenom: sourceChainTraceDenom,
		DelegateStrategy:      strategy,
		Authority:             authority,
	}
}

// GetSigners implements types.Msg
func (msg MsgAddSourceChain) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

func (msg MsgAddSourceChain) ValidateBasic() error {
	return nil
}

func NewMsgDelegate(
	chainID string,
	coin types.Coin,
	delegator string,
) *MsgDelegate {
	return &MsgDelegate{
		ChainId:   chainID,
		Coin:      coin,
		Delegator: delegator,
	}
}

func (msg MsgDelegate) ValidateBasic() error {
	return nil
}

// GetSigners implements types.Msg
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// GetSigners implements types.Msg
func (msg *MsgUnDelegate) GetSigners() []types.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// ValidateBasic implements types.Msg
func (*MsgUnDelegate) ValidateBasic() error {
	return nil
}

// GetSigners implements types.Msg
func (msg *MsgUpdateSourceChainDelegatePlan) GetSigners() []types.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// ValidateBasic implements types.Msg
func (*MsgUpdateSourceChainDelegatePlan) ValidateBasic() error {
	return nil
}

// GetSigners implements types.Msg
func (msg *MsgNotifyUnDelegationDone) GetSigners() []types.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

// ValidateBasic implements types.Msg
func (*MsgNotifyUnDelegationDone) ValidateBasic() error {
	return nil
}

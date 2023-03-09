package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgAddSourceChain{}
	_ sdk.Msg = &MsgDelegate{}
)

func NewMsgAddSourceChain(
	chainID string,
	connectionID string,
	version string,
	stakingDenom string,
	strategy []DelegationStrategy,
	authority string,
) *MsgAddSourceChain {
	return &MsgAddSourceChain{
		ChainId:          chainID,
		ConnectionId:     connectionID,
		Version:          version,
		StakingDenom:     stakingDenom,
		DelegateStrategy: strategy,
		Authority:        authority,
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

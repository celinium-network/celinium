package types

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// Callback type
const (
	DelegateTransferCall = iota
	DelegateCall
)

func (c *IBCCallback) CheckSuccessfulIBCAcknowledgement(cdc codec.Codec, responses []*codectypes.Any) bool {
	// TODO optimize the if/else code block
	if c.CallType == DelegateTransferCall {
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgSendResponse") {
				response := banktypes.MsgSendResponse{}
				err := cdc.Unmarshal(r.Value, &response)
				return err != nil
			}
		}
	} else if c.CallType == DelegateCall {
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgDelegateResponse") {
				response := stakingtypes.MsgDelegateResponse{}
				err := cdc.Unmarshal(r.Value, &response)
				return err != nil
			}
		}
	}
	return false
}

// func ()

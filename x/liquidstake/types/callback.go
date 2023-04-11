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
	UnbondCall
	WithdrawUnbondCall
)

// TODO don't check here
func (c *IBCCallback) CheckSuccessfulIBCAcknowledgement(cdc codec.Codec, responses []*codectypes.Any) bool {
	// TODO (1) optimize the if/else code block
	//      (2) check all response
	//      (3) should not be strings.Contains()
	switch c.CallType {
	case DelegateTransferCall:
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgSendResponse") {
				response := banktypes.MsgSendResponse{}
				err := cdc.Unmarshal(r.Value, &response)
				return err != nil
			}
		}
	case DelegateCall:
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgDelegateResponse") {
				return true
			}
		}
	case UnbondCall:
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgUndelegateResponse") {
				response := stakingtypes.MsgUndelegateResponse{}
				err := cdc.Unmarshal(r.Value, &response)
				return err != nil
			}
		}
	default:
		return false
	}

	return false
}

// func ()

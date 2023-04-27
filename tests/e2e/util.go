package e2e

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/codec/unknownproto"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	liquidstaketypes "github.com/celinium-network/celinium/x/liquidstake/types"
)

func decodeTx(cdc codec.Codec, txBytes []byte, interfaceRegs types.InterfaceRegistry) (*sdktx.Tx, error) {
	var raw sdktx.TxRaw

	// reject all unknown proto fields in the root TxRaw
	err := unknownproto.RejectUnknownFieldsStrict(txBytes, &raw, interfaceRegs)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(txBytes, &raw); err != nil {
		return nil, err
	}

	var body sdktx.TxBody
	if err := cdc.Unmarshal(raw.BodyBytes, &body); err != nil {
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	var authInfo sdktx.AuthInfo

	// reject all unknown proto fields in AuthInfo
	err = unknownproto.RejectUnknownFieldsStrict(raw.AuthInfoBytes, &authInfo, interfaceRegs)
	if err != nil {
		return nil, fmt.Errorf("failed to reject unknown fields: %w", err)
	}

	if err := cdc.Unmarshal(raw.AuthInfoBytes, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to decode auth info: %w", err)
	}

	return &sdktx.Tx{
		Body:       &body,
		AuthInfo:   &authInfo,
		Signatures: raw.Signatures,
	}, nil
}

func concatFlags(originalCollection []string, commandFlags []string, generalFlags []string) []string { //nolint:unused // this is called during e2e tests
	originalCollection = append(originalCollection, commandFlags...)
	originalCollection = append(originalCollection, generalFlags...)

	return originalCollection
}

func compareProxyDelegation(oriRecord, targetRecord *liquidstaketypes.ProxyDelegation) bool {
	if strings.Compare(oriRecord.ChainID, targetRecord.ChainID) != 0 {
		return false
	}
	if oriRecord.EpochNumber != targetRecord.EpochNumber {
		return false
	}
	if strings.Compare(oriRecord.Coin.Denom, targetRecord.Coin.GetDenom()) != 0 {
		return false
	}

	if !oriRecord.Coin.Amount.Equal(targetRecord.Coin.Amount) {
		return false
	}
	if oriRecord.Status != targetRecord.Status {
		return false
	}
	if !oriRecord.TransferredAmount.Equal(targetRecord.TransferredAmount) {
		return false
	}
	return true
}

package types

import (
	"bytes"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "inter-staking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName

	hostAccountsKey = "inter-account"
)

var (
	// Keys for store prefixes
	SourceChainMetadataKey   = []byte{0x11} // prefix for each key to a source chain metadata index
	SourceChainDelegationKey = []byte{0x12} // prefix for each key to a source chain total delegation index

	DelegationKey        = []byte{0x21} // key for a delegation
	DelegationQueueKey   = []byte{0x22} // key for delegation queue
	UndelegationQueueKey = []byte{0x23} // key for undelegation queue

	DistributionQueueKey = []byte{0x41} // key for distribution queue
)

// GetSourceChainMetadataKey return a key for a source chain whit chain id
func GetSourceChainMetadataKey(chainID []byte) []byte {
	return append(SourceChainMetadataKey, lengthPrefix(chainID)...)
}

func GetSourceChainDelegationKey(chainID []byte) []byte {
	return append(SourceChainDelegationKey, lengthPrefix(chainID)...)
}

func lengthPrefix(bz []byte) []byte {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}

func GetDelegateQueueKey(height uint64) []byte {
	heightBz := sdk.Uint64ToBigEndian(height)

	prefixL := len(DelegationQueueKey)

	bz := make([]byte, prefixL+8+8)

	copy(bz[:prefixL], DelegationQueueKey)
	copy(bz[prefixL:prefixL+8], heightBz)

	return bz
}

func ParseDelegateQueueKey(bz []byte) (uint64, error) {
	prefixL := len(DelegationQueueKey)

	if prefix := bz[:prefixL]; !bytes.Equal(prefix, DelegationQueueKey) {
		return 0, fmt.Errorf("invalid prefix; expected: %X, got: %X", DelegationQueueKey, prefix)
	}

	height := sdk.BigEndianToUint64(bz[prefixL : prefixL+8])

	return height, nil
}

package types

import (
	"bytes"
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "inter-staking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	// Keys for store prefixes
	SourceChainMetadataKey   = []byte{0x11} // prefix for each key to a source chain metadata index
	SourceChainDelegationKey = []byte{0x12} // prefix for each key to a source chain total delegation index

	// Delegation task status
	// Pending: a pending task, ICA get host chain native coin from delegator in control chain.
	// Preparing: ICA is transferring host chain coin back to hist chain
	// Prepared: ICA transfer back successfully.
	// Delegatiing: ICA is delegating on host chain.
	// Done: The delegation is successful.
	PendingDelegationQueueKey   = []byte{0x21} // key for a pending delegation queue
	PreparingDelegationQueueKey = []byte{0x22} // key for a prepare delegation queue
	PreparedDelegationQueueKey  = []byte{0x23} // key for a prepare delegation queue
	OngingDelegationQueueKey    = []byte{0x24} // key for a Ongoing delegation queue
	DelegationKey               = []byte{0x25} // key for a user delegation

	UndelegationQueueKey = []byte{0x23} // key for undelegation queue
	DistributionQueueKey = []byte{0x41} // key for distribution queue
)

var PercentageDenominator = math.NewIntFromUint64(100)

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

func GetDelegateQueueKey(queueKey []byte, height uint64) []byte {
	heightBz := sdk.Uint64ToBigEndian(height)

	prefixL := len(queueKey)

	bz := make([]byte, prefixL+8+8)

	copy(bz[:prefixL], queueKey)
	copy(bz[prefixL:prefixL+8], heightBz)

	return bz
}

func ParseDelegateQueueKey(bz []byte) (uint64, error) {
	prefixL := len(PendingDelegationQueueKey)

	if prefix := bz[:prefixL]; !bytes.Equal(prefix, PendingDelegationQueueKey) {
		return 0, fmt.Errorf("invalid prefix; expected: %X, got: %X", PendingDelegationQueueKey, prefix)
	}

	height := sdk.BigEndianToUint64(bz[prefixL : prefixL+8])

	return height, nil
}

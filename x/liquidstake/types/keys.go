package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "interstaking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the liquid stake module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

// Keys for store prefixes
var (
	EpochKey = []byte{0x10}

	// Prefix for source chain
	SouceChainKeyPrefix = []byte{0x11}

	// Key for delegation record ID.
	DelegationRecordIDKey = []byte{0x20}

	// Prefix for key which used in `{epoch + ChainID}=> DelegationRecordID`
	DelegationRecordIDForEpochPrefix = []byte{0x21}

	// Prefix for DelegationRecord `ID => DelegationRecord`
	DelegationRecordPrefix = []byte{0x22}

	// Prefix for key `{port + channel + sequence} => DelegationRecordID`
	IBCDelegationCallbackPrefix = []byte{0x23}
)

// GetSourceChainKey return key for source chain, `SouceChainKeyPrefix + len(chainID)+chainID`
func GetSourceChainKey(chainID []byte) []byte {
	return append(SouceChainKeyPrefix, lengthPrefix(chainID)...)
}

// GeChainDelegationRecordIDForEpochKey return , `SouceChainKeyPrefix + len(chainID)+chainID`
func GeChainDelegationRecordIDForEpochKey(epoch uint64, chainID []byte) []byte {
	epochBz := sdk.Uint64ToBigEndian(epoch)

	prefixL := len(DelegationRecordIDForEpochPrefix)

	chainIDWithLength := lengthPrefix(chainID)

	bz := make([]byte, prefixL+8+len(chainIDWithLength))

	copy(bz[:prefixL], DelegationRecordIDForEpochPrefix)
	copy(bz[prefixL:prefixL+8], epochBz)
	copy(bz[prefixL+8:], chainIDWithLength)

	return bz
}

func GetDelegationRecordKey(id uint64) []byte {
	idBz := sdk.Uint64ToBigEndian(id)

	return append(DelegationRecordPrefix, idBz...)
}

func GetIBCDelegationCallbackKey(channel uint64, sequence uint64) []byte {
	channelBz := sdk.Uint64ToBigEndian(channel)
	sequenceBz := sdk.Uint64ToBigEndian(sequence)

	prefixL := len(IBCDelegationCallbackPrefix)
	bz := make([]byte, +8+8)

	copy(bz[:prefixL], IBCDelegationCallbackPrefix)
	copy(bz[prefixL:prefixL+8], channelBz)
	copy(bz[prefixL+8:], sequenceBz)

	return bz
}

func lengthPrefix(bz []byte) []byte {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}

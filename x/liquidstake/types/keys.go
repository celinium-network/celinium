package types

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "liquidstake"

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
	ProxyDelegationIDKey = []byte{0x20}

	// Prefix for key which used in `{epoch + ChainID}=> ProxyDelegationID`
	ProxyDelegationIDPrefix = []byte{0x21}

	// Prefix for ProxyDelegation `ID => ProxyDelegation`
	ProxyDelegationPrefix = []byte{0x22}

	// Prefix for key `{channel + port + sequence} => ProxyDelegationID`
	IBCCallbackPrefix = []byte{0x23}

	// Prefix for key `{chainID + epoch + delegator}` => UnProxyDelegation
	UndelegationRecrodPrefix = []byte{0x31}

	EpochUnbondingsPrefix = []byte{0x32}

	IBCQueryKey = []byte{0x41}
)

// GetSourceChainKey return key for source chain, `SouceChainKeyPrefix + len(chainID)+chainID`
func GetSourceChainKey(chainID []byte) []byte {
	return append(SouceChainKeyPrefix, lengthPrefix(chainID)...)
}

// GetChainProxyDelegationIDForEpochKey return , `SouceChainKeyPrefix + len(chainID)+chainID`
func GetChainProxyDelegationIDForEpochKey(epoch uint64, chainID []byte) []byte {
	epochBz := sdk.Uint64ToBigEndian(epoch)

	prefixL := len(ProxyDelegationIDPrefix)

	chainIDWithLength := lengthPrefix(chainID)

	bz := make([]byte, prefixL+8+len(chainIDWithLength))

	copy(bz[:prefixL], ProxyDelegationIDPrefix)
	copy(bz[prefixL:prefixL+8], epochBz)
	copy(bz[prefixL+8:], chainIDWithLength)

	return bz
}

func GetProxyDelegationKey(id uint64) []byte {
	idBz := sdk.Uint64ToBigEndian(id)

	return append(ProxyDelegationPrefix, idBz...)
}

func GetIBCCallbackKey(channel []byte, port []byte, sequence uint64) []byte {
	channelBz := lengthPrefix(channel)
	portBz := lengthPrefix(port)
	sequenceBz := sdk.Uint64ToBigEndian(sequence)

	prefixL := len(IBCCallbackPrefix)
	channelBzL := len(channelBz)
	portBzL := len(portBz)

	bz := make([]byte, prefixL+channelBzL+portBzL+8)
	copy(bz[:prefixL], IBCCallbackPrefix)
	copy(bz[prefixL:prefixL+channelBzL], channelBz)
	copy(bz[prefixL+channelBzL:prefixL+channelBzL+portBzL], portBz)
	copy(bz[prefixL+channelBzL+portBzL:], sequenceBz)

	return bz
}

func GetUserUnbondingKey(chainID string, epoch uint64, delegator string) string {
	id := AssembleUserUnbondingID(chainID, epoch, delegator)

	return string(UndelegationRecrodPrefix) + id
}

func GetUndelegationRecordKeyFromID(id string) string {
	return string(UndelegationRecrodPrefix) + id
}

func AssembleUserUnbondingID(chainID string, epoch uint64, delegator string) string {
	return strings.Join([]string{chainID, strconv.FormatUint(epoch, 10), delegator}, ".")
}

func GetEpochUnbondingsKey(epoch uint64) []byte {
	be := sdk.Uint64ToBigEndian(epoch)

	return append(EpochUnbondingsPrefix, be...)
}

func lengthPrefix(bz []byte) []byte {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}

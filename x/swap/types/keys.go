package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "swap"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_swap"
)

var (
	// KeyNextGlobalPairId defines key to store the next Pool ID to be used.
	KeyNextGlobalPairId = []byte{0x01}

	KeyPrefixPools = []byte{0x02}

	KeyPrefixTokensToPoolId = []byte{0x03}
)

const KeySeparator = "|"

func GetKeyPrefixPairs(poolId uint64) []byte {
	return append(KeyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}

func GetKeyPrefixTokensToPoolId(token0, token1 string) []byte {
	tokenPrefix := []byte(strings.Join([]string{token0, token1}, KeySeparator))
	return append(KeyPrefixTokensToPoolId, tokenPrefix...)
}

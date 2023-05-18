package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/utils"
)

const (
	// ModuleName is the name of the multistaking module
	ModuleName = "multistaking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the multistaking module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	// Key for the denom white list which allow used for multistaking
	MultiStakingDenomWhiteListKey = []byte{0x11}

	// Prefix for key which used in `{denom + validator_address} => MultiStakingAgent's ID`
	MultiStakingAgentIDPrefix = []byte{0x21}

	// Prefix for Key which used in `id => MultiStakingAgent`
	MultiStakingAgentPrefix = []byte{0x22}

	MultiStakingLatestAgentIDKey = []byte{0x23}

	// Prefix for key which used in `{agent_id + delegator_address} => MultiStakingUnbonding`
	MultiStakingUnbondingPrefix = []byte{0x31}

	// Prefix for key which used in `{agent_id + delegator_address} => shares_amount`
	MultiStakingSharesPrefix = []byte{0x41}
)

func GetMultiStakingAgentIDKey(denom, valAddr string) []byte {
	denomBz := utils.BytesLengthPrefix([]byte(denom))
	valAddrBz := utils.BytesLengthPrefix([]byte(valAddr))

	prefixLen := len(MultiStakingAgentIDPrefix)
	denomBzLen := len(denomBz)
	valAddrBzLen := len(valAddrBz)

	bz := make([]byte, prefixLen+denomBzLen+valAddrBzLen)

	copy(bz[:prefixLen], MultiStakingAgentIDPrefix)
	copy(bz[prefixLen:prefixLen+denomBzLen], denomBz)
	copy(bz[prefixLen+denomBzLen:], valAddrBz)

	return bz
}

func GetMultiStakingAgentKey(agentID uint64) []byte {
	idBz := sdk.Uint64ToBigEndian(agentID)
	return append(MultiStakingAgentIDPrefix, idBz...)
}

func GetMultiStakingSharesKey(agentID uint64, delegator string) []byte {
	idBz := sdk.Uint64ToBigEndian(agentID)
	delegatorBz := utils.BytesLengthPrefix([]byte(delegator))
	prefixLen := len(MultiStakingSharesPrefix)

	bz := make([]byte, prefixLen+8+len(delegatorBz))
	copy(bz[:prefixLen], MultiStakingSharesPrefix)
	copy(bz[prefixLen:prefixLen+8], idBz)
	copy(bz[prefixLen+8:], delegatorBz)

	return bz
}

func GetMultiStakingUnbondingKey(agentID uint64, delegator string) []byte {
	idBz := sdk.Uint64ToBigEndian(agentID)
	delegatorBz := utils.BytesLengthPrefix([]byte(delegator))
	prefixLen := len(MultiStakingUnbondingPrefix)

	bz := make([]byte, prefixLen+8+len(delegatorBz))

	copy(bz[:prefixLen], MultiStakingUnbondingPrefix)
	copy(bz[prefixLen:prefixLen+8], idBz)
	copy(bz[prefixLen+8:], delegatorBz)

	return bz
}

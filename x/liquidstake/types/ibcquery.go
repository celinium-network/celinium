package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (q IBCQuery) ID(blockHeight uint64) string {
	id := append(sdk.Uint64ToBigEndian(blockHeight), []byte(q.ChainID+q.ConnectionID+q.QueryType+q.QueryPathKey)...)
	return fmt.Sprintf("%x", id)
}

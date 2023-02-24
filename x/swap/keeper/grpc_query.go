package keeper

import (
	context "context"

	"celinium/x/swap/types"
)

var _ types.QueryServer = Keeper{}

// PairByTokens implements types.QueryServer
func (Keeper) PairByTokens(context.Context, *types.QueryPairByTokensRequest) (*types.QueryPairByTokensReponse, error) {
	panic("unimplemented")
}

// Params implements types.QueryServer
func (Keeper) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	panic("unimplemented")
}

package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// ProxyDelegation implements types.QueryServer
func (k Querier) ProxyDelegation(goCtx context.Context, req *types.QueryProxyDelegationRequest) (*types.QueryProxyDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ChainID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty chainID %s", req.ChainID)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := k.GetSourceChain(ctx, req.ChainID); !found {
		return nil, status.Errorf(codes.InvalidArgument, "unknown chainID %s", req.ChainID)
	}

	recordID, found := k.GetChianProxyDelegationID(ctx, req.ChainID, req.Epoch)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "chain%s no record in epoch %d", req.ChainID, req.Epoch)
	}

	record, found := k.GetProxyDelegation(ctx, recordID)
	if !found {
		return nil, status.Errorf(codes.Internal, "chain%s no record in epoch %d", req.ChainID, req.Epoch)
	}
	return &types.QueryProxyDelegationResponse{
		Record: *record,
	}, nil
}

// EpochProxyUnbonding implements types.QueryServer
func (k Querier) EpochProxyUnbonding(goCtx context.Context, req *types.QueryEpochProxyUnbondingRequest) (*types.QueryEpochProxyUnbondingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ChainID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty chainID %s", req.ChainID)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	unbondings, found := k.GetEpochUnboundings(ctx, req.Epoch)
	if !found || len(unbondings.Unbondings) == 0 {
		return nil, status.Errorf(codes.Internal, "no unbondings in epoch %d", req.Epoch)
	}

	var chainUnbonding types.ProxyUnbonding
	for _, unbonding := range unbondings.Unbondings {
		if strings.Compare(req.ChainID, unbonding.ChainID) != 0 {
			continue
		}
		chainUnbonding = unbonding
	}
	return &types.QueryEpochProxyUnbondingResponse{
		ChainUnbonding: chainUnbonding,
	}, nil
}

// SourceChain implements types.QueryServer
func (k Querier) SourceChain(goCtx context.Context, req *types.QuerySourceChainRequest) (*types.QuerySourceChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ChainID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "empty chainID %s", req.ChainID)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	sourceChain, found := k.GetSourceChain(ctx, req.ChainID)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "unknown chainID %s", req.ChainID)
	}

	return &types.QuerySourceChainResponse{
		SourceChain: *sourceChain,
	}, nil
}

// UserUnbonding implements types.QueryServer
func (k Querier) UserUnbonding(goCtx context.Context, req *types.QueryUserUnbondingRequest) (*types.QueryUserUnbondingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := k.GetSourceChain(ctx, req.ChainID)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "unknown chainID %s", req.ChainID)
	}

	curEpoch, found := k.epochKeeper.GetEpochInfo(ctx, appparams.UndelegationEpochIdentifier)

	if !found {
		return nil, status.Errorf(codes.Internal, "undelegation epoch not start")
	}

	var userUnbondings []types.UserUnbonding
	// TODO the loop maybe expensive. so get epoch from request?
	for i := uint64(0); i < uint64(curEpoch.CurrentEpoch); i++ {
		record, found := k.GetUndelegationRecord(ctx, req.ChainID, i, req.User)
		if !found {
			continue
		}
		userUnbondings = append(userUnbondings, *record)
	}

	return &types.QueryUserUnbondingResponse{
		UserUnbondings: userUnbondings,
	}, nil
}

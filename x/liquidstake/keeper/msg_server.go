package keeper

import (
	ctx "context"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the interstaking Msg service interface
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

type msgServer struct {
	*Keeper
}

// RegisterSourceChain implements types.MsgServer
func (msgServer) RegisterSourceChain(ctx.Context, *types.MsgRegisterSourceChain) (*types.MsgRegisterSourceChainResponse, error) {
	panic("unimplemented")
}

// EditValidators implements types.MsgServer
func (msgServer) EditValidators(ctx.Context, *types.MsgEditVadlidators) (*types.MsgEditValidatorsResponse, error) {
	panic("unimplemented")
}

// RebalanceValidator implements types.MsgServer
func (msgServer) RebalanceValidator(ctx.Context, *types.MsgRebalanceValidators) (*types.RebalanceValidatorsResponse, error) {
	panic("unimplemented")
}

// Delegate implements types.MsgServer
func (msgServer) Delegate(ctx.Context, *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	panic("unimplemented")
}

// Undelegate implements types.MsgServer
func (msgServer) Undelegate(ctx.Context, *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	panic("unimplemented")
}

// Claim implements types.MsgServer
func (msgServer) Claim(ctx.Context, *types.MsgClaim) (*types.MsgClaimResponse, error) {
	panic("unimplemented")
}

// Reinvest implements types.MsgServer
func (msgServer) Reinvest(ctx.Context, *types.MsgReinvest) (*types.MsgReinvestResponse, error) {
	panic("unimplemented")
}

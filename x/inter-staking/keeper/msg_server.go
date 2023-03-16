package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/inter-staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the interstaking Msg service interface
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// AddSourceChain implements types.MsgServer
func (m msgServer) AddSourceChain(goCtx context.Context, msg *types.MsgAddSourceChain) (*types.MsgAddSourceChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check singer
	// if msg.Authority != m.Keeper.authority {
	// 	return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	// }

	icaCtlAddr, err := m.Keeper.AddSourceChain(
		ctx, msg.DelegateStrategy,
		msg.SourceChainDenom,
		msg.SourceChainTraceDenom,
		msg.ChainId,
		msg.ConnectionId,
		msg.ChannelId,
		msg.Version)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddSourceChainResponse{
		InterchainAccount: icaCtlAddr,
	}, nil
}

// Delegate implements types.MsgServer
//
// Delegate will create a delegation task and push it to queue.
// Pop a certain number of tasks at `EndBlock` and assemble them
// into a cross-chain transaction.
func (m msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.Delegate(ctx, msg.ChainId, msg.Coin, msg.Delegator); err != nil {
		return nil, err
	}
	// Emit Event
	return &types.MsgDelegateResponse{}, nil
}

// NotifyUnDelegationDone implements types.MsgServer
func (msgServer) NotifyUnDelegationDone(context.Context, *types.MsgNotifyUnDelegationDone) (*types.MsgNotifyDelegationDoneResponse, error) {
	panic("unimplemented")
}

// Undelegate implements types.MsgServer
func (msgServer) Undelegate(context.Context, *types.MsgUnDelegate) (*types.MsgUnDelegateResponse, error) {
	panic("unimplemented")
}

// UpdateSourceChainDelegatePlan implements types.MsgServer
func (msgServer) UpdateSourceChainDelegatePlan(context.Context, *types.MsgUpdateSourceChainDelegatePlan) (*types.MsgUpdateSourceChainDelegatePlanResponse, error) {
	panic("unimplemented")
}

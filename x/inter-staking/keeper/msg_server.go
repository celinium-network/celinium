package keeper

import (
	context "context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

	if len(msg.DelegateStrategy) == 0 {
		return nil, sdkerrors.Wrapf(types.ErrMismatchParameter, "the delegate plan should be set")
	}

	if exist, _ := m.SourceChainExist(ctx, msg.ChainId); exist {
		return nil, sdkerrors.Wrapf(types.ErrSourceChainExist, "source chain: %s already exist", msg.ChainId)
	}

	icaCtlAccount := types.GenerateSourceChainControlAccount(ctx, msg.ChainId, msg.ConnectionId)

	m.accountKeeper.NewAccount(ctx, icaCtlAccount)
	m.accountKeeper.SetAccount(ctx, icaCtlAccount)

	icaCtladdr := icaCtlAccount.GetAddress().String()

	if err := m.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, icaCtladdr, msg.Version); err != nil {
		return nil, err
	}

	sourceChainMetaData := types.SourceChainMetadata{
		IbcClientId:      msg.ChainId,
		IbcConnectionId:  msg.ConnectionId,
		ICAControlAddr:   icaCtladdr,
		StakingDenom:     msg.StakingDenom,
		DelegateStrategy: msg.DelegateStrategy,
	}

	m.SetSourceChain(ctx, msg.ChainId, &sourceChainMetaData)

	return &types.MsgAddSourceChainResponse{
		InterchainAccount: icaCtladdr,
	}, nil
}

// Delegate implements types.MsgServer
//
// Delegate will create a delegation task and push it to queue.
// Pop a certain number of tasks at `EndBlock` and assemble them
// into a cross-chain transaction.
func (m msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sourceChainMetadata, found := m.GetSourceChain(ctx, msg.ChainId)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", msg.ChainId)
	}

	// check wheather the ica of source chain in control endpoint is available.
	if !m.SourceChainAvaiable(ctx, sourceChainMetadata.IbcConnectionId, sourceChainMetadata.ICAControlAddr) {
		return nil, sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", msg.ChainId)
	}

	// Check wheather the coin is the native token of the source chain.
	if strings.Compare(msg.Coin.Denom, sourceChainMetadata.StakingDenom) != 0 {
		return nil, sdkerrors.Wrapf(types.ErrMismatchSourceCoin, "chainID: %s, expected: %s, get:",
			msg.ChainId, sourceChainMetadata.StakingDenom, msg.Coin.Denom)
	}

	if err := m.SendCoinsFromDelegatorToICA(ctx, msg.Delegator, sourceChainMetadata.ICAControlAddr, sdk.Coins{msg.Coin}); err != nil {
		return nil, err
	}

	newDelegationTask := types.DelegationTask{
		ChainId:   msg.ChainId,
		Delegator: msg.Delegator,
		Amount:    msg.Coin,
	}

	m.PushDelegationTaskQueue(&ctx, types.PendingDelegationQueueKey, &newDelegationTask)

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

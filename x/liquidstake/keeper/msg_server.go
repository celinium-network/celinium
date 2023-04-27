package keeper

import (
	goctx "context"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the interstaking Msg service interface
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper *Keeper
}

// RegisterSourceChain implements types.MsgServer
func (ms msgServer) RegisterSourceChain(goCtx goctx.Context, msg *types.MsgRegisterSourceChain) (*types.MsgRegisterSourceChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sourceChain := types.SourceChain{
		ChainID:                   msg.ChainID,
		ConnectionID:              msg.ConnectionID,
		TransferChannelID:         msg.TrasnferChannelID,
		Bech32ValidatorAddrPrefix: msg.Bech32ValidatorAddrPrefix,
		Validators:                msg.Validators,
		Redemptionratio:           sdk.NewDecWithPrec(100000000, 8),
		NativeDenom:               msg.NativeDenom,
		DerivativeDenom:           msg.DerivativeDenom,
		StakedAmount:              math.ZeroInt(),
	}

	if err := ms.keeper.AddSouceChain(ctx, &sourceChain); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterSourceChain,
			sdk.NewAttribute(types.AttributeKeySourceChainID, msg.ChainID),
			sdk.NewAttribute(types.AttributeKeyValidators, sourceChain.ValidatorsAddress()),
			sdk.NewAttribute(types.AttributeKeyWeights, sourceChain.ValidatorsWeight()),
		),
	)

	return &types.MsgRegisterSourceChainResponse{}, nil
}

// EditValidators implements types.MsgServer
func (msgServer) EditValidators(goctx.Context, *types.MsgEditVadlidators) (*types.MsgEditValidatorsResponse, error) {
	panic("unimplemented")
}

// RebalanceValidator implements types.MsgServer
func (msgServer) RebalanceValidator(goctx.Context, *types.MsgRebalanceValidators) (*types.RebalanceValidatorsResponse, error) {
	panic("unimplemented")
}

// Delegate implements types.MsgServer
func (ms msgServer) Delegate(goCtx goctx.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddress := sdk.MustAccAddressFromBech32(msg.Delegator)

	record, err := ms.keeper.Delegate(ctx, msg.ChainID, msg.Amount, delegatorAccAddress)
	if err != nil {
		return nil, nil
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.Delegator),
			sdk.NewAttribute(types.AttributeKeySourceChainID, msg.ChainID),
			sdk.NewAttribute(types.AttributeKeyEpoch, strconv.FormatUint(record.EpochNumber, 10)),
			sdk.NewAttribute(types.AttributeKeyDelegateAmt, msg.Amount.String()),
		),
	)

	return &types.MsgDelegateResponse{}, nil
}

// Undelegate implements types.MsgServer
func (ms msgServer) Undelegate(goCtx goctx.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddress := sdk.MustAccAddressFromBech32(msg.Delegator)
	record, err := ms.keeper.Undelegate(ctx, msg.ChainID, msg.Amount, delegatorAccAddress)
	if err != nil {
		return nil, nil
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUndelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.Delegator),
			sdk.NewAttribute(types.AttributeKeySourceChainID, msg.ChainID),
			sdk.NewAttribute(types.AttributeKeyEpoch, strconv.FormatUint(record.Epoch, 10)),
			sdk.NewAttribute(types.AttributeKeyUnbondAmt, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyRedeemAmt, record.RedeemCoin.Amount.String()),
		),
	)
	return &types.MsgUndelegateResponse{}, nil
}

// Claim implements types.MsgServer
func (ms msgServer) Claim(goCtx goctx.Context, msg *types.MsgClaim) (*types.MsgClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddress := sdk.MustAccAddressFromBech32(msg.Delegator)
	claimAmt, err := ms.keeper.ClaimUnbonding(ctx, delegatorAccAddress, msg.Epoch, msg.ChainId)
	if err != nil {
		return nil, nil
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUndelegate,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.Delegator),
			sdk.NewAttribute(types.AttributeKeySourceChainID, msg.ChainId),
			sdk.NewAttribute(types.AttributeKeyEpoch, strconv.FormatUint(msg.Epoch, 10)),
			sdk.NewAttribute(types.AttributeKeyClaimAmt, claimAmt.String()),
		),
	)

	return nil, nil
}

// Reinvest implements types.MsgServer
func (msgServer) Reinvest(goctx.Context, *types.MsgReinvest) (*types.MsgReinvestResponse, error) {
	panic("unimplemented")
}

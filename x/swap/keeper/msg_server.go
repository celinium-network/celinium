package keeper

import (
	context "context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/swap/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreatePair implements types.MsgServer
func (m msgServer) CreatePair(goCtx context.Context, msg *types.MsgCreatePair) (*types.MsgCreateaPairResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pair, err := m.Keeper.createPair(ctx, msg.Token0, msg.Token1)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgCreatePair,
			sdk.NewAttribute(types.AttributeCaller, msg.Sender),
			sdk.NewAttribute(types.AttributeToken0, pair.Token0.Denom),
			sdk.NewAttribute(types.AttributeToken0, pair.Token1.Denom),
			sdk.NewAttribute(types.AttributePairAccount, pair.Account),
		),
	})

	return &types.MsgCreateaPairResponse{
		NewLpToken: pair.LpToken.Denom,
	}, nil
}

// AddLiquidity implements types.MsgServer
func (m msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	var err error
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	res, err := m.Keeper.addLiquidty(ctx, senderAddr, &msg.Token0, &msg.Token1, msg.Amount0Min, msg.Amount1Min, senderAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgAddLiquidty,
			sdk.NewAttribute(types.AttributeCaller, msg.Sender),
			sdk.NewAttribute(types.AttributeToken0, res.Token0.Denom),
			sdk.NewAttribute(types.AttributeToken0Amount, res.Token0.Amount.String()),
			sdk.NewAttribute(types.AttributeToken1, res.Token1.Denom),
			sdk.NewAttribute(types.AttributeToken1Amount, res.Token1.Amount.String()),
			sdk.NewAttribute(types.AttributeLPToken, res.TpToken.Denom),
			sdk.NewAttribute(types.AttributeLPTokenAmount, res.TpToken.Amount.String()),
		),
	})

	return &types.MsgAddLiquidityResponse{
		Sender:  msg.Sender,
		Token0:  *res.Token0,
		Token1:  *res.Token1,
		LpToken: *res.TpToken,
	}, nil
}

// SwapExactTokensForTokens implements types.MsgServer
func (m msgServer) SwapExactTokensForTokens(goCtx context.Context, msg *types.MsgSwapExactTokensForTokens) (*types.MsgSwapExactTokensForTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	amounts, err := m.Keeper.swapExactTokensForTokens(ctx, senderAddr, msg.AmountIn, msg.AmountOutMin, msg.Path, recipient)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeMsgSwapExactTokensForTokens,
			sdk.NewAttribute(types.AttributeCaller, msg.Sender),
			sdk.NewAttribute(types.AttributePath, strings.Join(msg.Path, ",")),
			sdk.NewAttribute(types.AttributeAmountIn, msg.AmountIn.String()),
			sdk.NewAttribute(types.AttributeAmountOut, amounts[len(amounts)-1].String()),
			sdk.NewAttribute(types.AttributeRecipient, msg.Recipient),
		),
	})

	return &types.MsgSwapExactTokensForTokensResponse{
		Sender:    msg.Sender,
		Recipient: msg.Recipient,
		Path:      msg.Path,
		Ammounts:  amounts,
	}, nil
}

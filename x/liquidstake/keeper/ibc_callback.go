package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (k Keeper) SetCallBack(ctx sdk.Context, channel string, port string, sequence uint64, callback *types.IBCCallback) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(callback)
	store.Set(types.GetIBCCallbackKey([]byte(channel), []byte(port), sequence), bz)
}

func (k Keeper) GetCallBack(ctx sdk.Context, channel string, port string, sequence uint64) (*types.IBCCallback, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetIBCCallbackKey([]byte(channel), []byte(port), sequence))
	if bz == nil {
		return nil, false
	}

	callback := types.IBCCallback{}
	k.cdc.MustUnmarshal(bz, &callback)

	return &callback, true
}

func GetResultFromAcknowledgement(acknowledgement []byte) ([]byte, error) {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return nil, err
	}

	switch response := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		if len(response.Result) == 0 {
			return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "empty acknowledgement")
		}
		return ack.GetResult(), nil
	case *channeltypes.Acknowledgement_Error:
		return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidPacket, "invalid acknowledgement")
	default:
		return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "unknown acknowledgement status")

	}
}

func (k Keeper) HandleIBCTransferAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, acknowledgement []byte) error {
	_, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("IBC acknowledgement has error %v", err))
		return nil
	}

	callback, found := k.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("callback not exit, channelID: %s, portID: %s, sequence: %d",
			packet.SourceChannel, packet.SourcePort, packet.Sequence))
		return nil
	}

	handler, ok := callbackHandlerRegistry[callback.CallType]
	if !ok {
		return nil
	}

	return handler(&k, ctx, callback, nil)
	// TODO consider remove callback ?, repeated receive same Acknowledgement
}

func (k Keeper) HandleICAAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("IBC acknowledgement has error %v", err))
		return nil
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	callback, found := k.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("callback not exit, channelID: %s, portID: %s, sequence: %d",
			packet.SourceChannel, packet.SourcePort, packet.Sequence))
		return nil
	}

	handler, ok := callbackHandlerRegistry[callback.CallType]
	if !ok {
		return nil
	}

	return handler(&k, ctx, callback, txMsgData.MsgResponses)

	// TODO consider remove callback ?, repeated receive same Acknowledgement
}

package interstaking

import (
	proto "github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"celinium/x/inter-staking/keeper"
)

var _ porttypes.IBCModule = IBCModule{}

// / IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnAcknowledgementPacket implements types.IBCModule
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 packet acknowledgement: %v", err)
	}

	var txMsgData sdk.TxMsgData
	if err := proto.Unmarshal(ack.GetResult(), &txMsgData); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-27 tx message data: %v", err)
	}

	if len(txMsgData.Data) == 0 {
		for _, msgResp := range txMsgData.GetMsgResponses() {
			// im.keeper.Logger(ctx).Info("msg response in ICS-27 packet", "response", msgResp.GoString(), "typeURL", msgResp.GetTypeUrl())
			delegateResponse := &stakingtypes.MsgDelegateResponse{}

			err := proto.Unmarshal(msgResp.Value, delegateResponse)
			if err == nil {
				im.keeper.OnAcknowledgement(ctx, &packet)
			}
		}
	} else {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "inter staking can't process crosschain tx")
	}
	return nil
}

// OnChanCloseConfirm implements types.IBCModule
func (IBCModule) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanCloseInit implements types.IBCModule
func (IBCModule) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanOpenAck implements types.IBCModule
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return nil
}

// OnChanOpenConfirm implements types.IBCModule
func (IBCModule) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanOpenInit implements types.IBCModule
func (IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return "", nil
}

// OnChanOpenTry implements types.IBCModule
func (IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return "", nil
}

// OnRecvPacket implements types.IBCModule
func (IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	return nil
}

// OnTimeoutPacket implements types.IBCModule
func (IBCModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return nil
}

package liquidstake

import (
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"celinium/x/liquidstake/keeper"
)

var _ porttypes.IBCModule = IBCModule{}

// / IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
	cdc    codec.Codec
}

// OnAcknowledgementPacket implements types.IBCModule
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := im.cdc.Unmarshal(ack.GetResult(), &txMsgData); err != nil {
		return err
	}

	return im.keeper.HandleIBCAcknowledgement(ctx, &packet, txMsgData.MsgResponses)
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
func (IBCModule) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return nil
}

// OnChanOpenConfirm implements types.IBCModule
func (IBCModule) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanOpenInit implements types.IBCModule
func (IBCModule) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return "", nil
}

// OnChanOpenTry implements types.IBCModule
func (IBCModule) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
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

func NewIBCModule(k keeper.Keeper, cdc codec.Codec) IBCModule {
	return IBCModule{
		keeper: k,
		cdc:    cdc,
	}
}

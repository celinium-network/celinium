package liquidstake

import (
	"celinium/x/liquid-stake/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

var _ porttypes.IBCModule = IBCModule{}

// / IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// OnAcknowledgementPacket implements types.IBCModule
func (IBCModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	panic("unimplemented")
}

// OnChanCloseConfirm implements types.IBCModule
func (IBCModule) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanCloseInit implements types.IBCModule
func (IBCModule) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenAck implements types.IBCModule
func (IBCModule) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	panic("unimplemented")
}

// OnChanOpenConfirm implements types.IBCModule
func (IBCModule) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenInit implements types.IBCModule
func (IBCModule) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	panic("unimplemented")
}

// OnChanOpenTry implements types.IBCModule
func (IBCModule) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	panic("unimplemented")
}

// OnRecvPacket implements types.IBCModule
func (IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	panic("unimplemented")
}

// OnTimeoutPacket implements types.IBCModule
func (IBCModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	panic("unimplemented")
}

func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

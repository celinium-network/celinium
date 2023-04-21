package liquidstake

import (
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/celinium-network/celinium/x/liquidstake/keeper"
)

var _ porttypes.IBCModule = IBCModule{}

// / IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
	cdc    codec.Codec
}

// OnAcknowledgementPacket implements types.IBCModule
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.keeper.HandleICAAcknowledgement(ctx, &packet, acknowledgement)
}

// OnChanCloseConfirm implements types.IBCModule
func (IBCModule) OnChanCloseConfirm(sdk.Context, string, string) error {
	return nil
}

// OnChanCloseInit implements types.IBCModule
func (IBCModule) OnChanCloseInit(sdk.Context, string, string) error {
	return nil
}

// OnChanOpenAck implements types.IBCModule
func (IBCModule) OnChanOpenAck(sdk.Context, string, string, string, string) error {
	return nil
}

// OnChanOpenConfirm implements types.IBCModule
func (IBCModule) OnChanOpenConfirm(sdk.Context, string, string) error {
	return nil
}

// OnChanOpenInit implements types.IBCModule
func (IBCModule) OnChanOpenInit(sdk.Context, channeltypes.Order, []string, string, string, *capabilitytypes.Capability, channeltypes.Counterparty, string) (string, error) {
	return "", nil
}

// OnChanOpenTry implements types.IBCModule
func (IBCModule) OnChanOpenTry(sdk.Context, channeltypes.Order, []string, string, string, *capabilitytypes.Capability, channeltypes.Counterparty, string) (version string, err error) {
	return "", nil
}

// OnRecvPacket implements types.IBCModule
func (IBCModule) OnRecvPacket(sdk.Context, channeltypes.Packet, sdk.AccAddress) exported.Acknowledgement {
	return nil
}

// OnTimeoutPacket implements types.IBCModule
func (IBCModule) OnTimeoutPacket(sdk.Context, channeltypes.Packet, sdk.AccAddress) error {
	return nil
}

func NewIBCModule(k keeper.Keeper, cdc codec.Codec) IBCModule {
	return IBCModule{
		keeper: k,
		cdc:    cdc,
	}
}

package coordinator

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/celinium-network/celinium/x/restaking/coordinator/keeper"
	restaking "github.com/celinium-network/celinium/x/restaking/types"
)

var _ porttypes.IBCModule = IBCModule{}

// / IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper *keeper.Keeper
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
	return version, errorsmod.Wrap(restaking.ErrInvalidChannelFlow, "channel handshake must be initiated by consumer chain")
}

// OnChanOpenTry implements types.IBCModule
func (am IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	// Validate parameters
	if err := validateCCVChannelParams(
		ctx, am.keeper, order, portID,
	); err != nil {
		return "", err
	}

	if counterparty.PortId != restaking.ConsumerPortID {
		return "", errorsmod.Wrapf(porttypes.ErrInvalidPort,
			"invalid counterparty port: %s, expected %s", counterparty.PortId, restaking.ConsumerPortID)
	}

	cpv, err := am.keeper.ParseCounterPartyVersion(counterpartyVersion)
	if err != nil {
		return "", err
	}
	if cpv.Version != restaking.Version {
		return "", errorsmod.Wrapf(
			restaking.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s",
			counterpartyVersion, restaking.Version)
	}

	if err := am.keeper.ClaimCapability(
		ctx, channelCap, host.ChannelCapabilityPath(portID, channelID),
	); err != nil {
		return "", err
	}

	if len(connectionHops) != 1 {
		return "", errorsmod.Wrap(channeltypes.ErrTooManyConnectionHops, "must have direct connection to provider chain")
	}

	connectionID := connectionHops[0]
	clientID, tmClient, err := am.keeper.GetUnderlyingClient(ctx, connectionID)
	if err != nil {
		return "", err
	}

	if err := am.keeper.VerifyConnectingConsumer(ctx, tmClient); err != nil {
		return "", err
	}

	if err := am.keeper.VerifyConsumerValidatorSet(ctx, clientID, cpv.ValidatorSet); err != nil {
		return "", err
	}

	am.keeper.SetConsumerClientID(ctx, tmClient.ChainId, clientID)
	am.keeper.SetConsumerClientValidatorSet(ctx, tmClient.ChainId, cpv.ValidatorSet)

	// TODO remove consumer additional proposal?

	return "", nil
}

// OnChanOpenAck implements types.IBCModule
func (IBCModule) OnChanOpenAck(
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
	panic("unimplemented")
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

// OnRecvPacket implements types.IBCModule
func (IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	panic("unimplemented")
}

// OnTimeoutPacket implements types.IBCModule
func (IBCModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	panic("unimplemented")
}

// validateCCVChannelParams validates a ccv channel
func validateCCVChannelParams(
	ctx sdk.Context,
	keeper *keeper.Keeper,
	order channeltypes.Order,
	portID string,
) error {
	if order != channeltypes.ORDERED {
		return errorsmod.Wrapf(channeltypes.ErrInvalidChannelOrdering, "expected %s channel, got %s ", channeltypes.ORDERED, order)
	}

	// the port ID must match the port ID the CCV module is bounded to
	boundPort := keeper.GetPort(ctx)
	if boundPort != portID {
		return errorsmod.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}
	return nil
}

package keeper

import (
	"bytes"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/celinium-network/celinium/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/celinium/x/restaking/types"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.Codec
	paramSpace paramtypes.Subspace

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper restaking.IBCTransferKeeper
	ibcCoreKeeper     restaking.IBCCoreKeeper

	scopedKeeper restaking.ScopedKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper restaking.IBCTransferKeeper,
	ibcCoreKeeper restaking.IBCCoreKeeper,
	scopedKeeper restaking.ScopedKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	// if !paramSpace.HasKeyTable() {
	// 		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	// }

	k := Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		channelKeeper:     channelKeeper,
		portKeeper:        portKeeper,
		connectionKeeper:  connectionKeeper,
		clientKeeper:      clientKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
		ibcCoreKeeper:     ibcCoreKeeper,
		scopedKeeper:      scopedKeeper,
	}

	return k
}

func (k Keeper) SetConsumerAdditionProposal(ctx sdk.Context, prop *types.ConsumerAdditionProposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(prop)
	store.Set(types.ConsumerAdditionProposalKey(prop.ChainId), bz)
}

func (k Keeper) GetConsumerClientID(ctx sdk.Context, chainID string) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerClientIDKey(chainID))

	return bz, bz != nil
}

func (k Keeper) SetConsumerClientID(ctx sdk.Context, chainID, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsumerClientIDKey(chainID), []byte(clientID))
}

func (k Keeper) SetConsumerClientValidatorSet(ctx sdk.Context, chainID string, valSet restaking.ValidatorSet) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&valSet)
	store.Set(types.ConsumerValidatorSetKey(chainID), bz)
}

func (k Keeper) SetPendingConsumerAdditionProposal(ctx sdk.Context, prop *types.ConsumerAdditionProposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(prop)
	store.Set(types.ConsumerAdditionProposalKey(prop.ChainId), bz)
}

func (k Keeper) GetPendingConsumerAdditionProposal(ctx sdk.Context, chainID string) (*types.ConsumerAdditionProposal, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ConsumerAdditionProposalKey(chainID))
	if bz == nil {
		return nil, false
	}

	var prop types.ConsumerAdditionProposal
	if err := k.cdc.Unmarshal(bz, &prop); err != nil {
		return nil, false
	}

	return &prop, true
}

// GetPort returns the portID for the CCV module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey()))
}

// SetPort sets the portID for the CCV module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey(), []byte(portID))
}

func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// VerifyConnectingConsumer
func (k Keeper) VerifyConnectingConsumer(ctx sdk.Context, tmClient *ibctmtypes.ClientState) error {
	prop, found := k.GetPendingConsumerAdditionProposal(ctx, tmClient.ChainId)
	if !found {
		return errorsmod.Wrapf(restaking.ErrUnauthorizedConsumerChain,
			"Consumer chain is not authorized ChainID:%s", tmClient.ChainId)
	}

	if err := verifyConsumerAdditionProposal(prop, tmClient); err != nil {
		return err
	}

	return nil
}

// Retrieves the underlying client state corresponding to a connection ID.
func (k Keeper) GetUnderlyingClient(ctx sdk.Context, connectionID string) (
	clientID string, tmClient *ibctmtypes.ClientState, err error,
) {
	conn, ok := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !ok {
		return "", nil, errorsmod.Wrapf(conntypes.ErrConnectionNotFound,
			"connection not found for connection ID: %s", connectionID)
	}
	clientID = conn.ClientId
	clientState, ok := k.clientKeeper.GetClientState(ctx, clientID)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrClientNotFound,
			"client not found for client ID: %s", conn.ClientId)
	}
	tmClient, ok = clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrInvalidClientType,
			"invalid client type. expected %s, got %s", ibcexported.Tendermint, clientState.ClientType())
	}
	return clientID, tmClient, nil
}

func (k Keeper) ParseCounterPartyVersion(version string) (*restaking.CounterPartyVersion, error) {
	var cpv restaking.CounterPartyVersion
	err := k.cdc.Unmarshal([]byte(version), &cpv)
	if err != nil {
		return nil, err
	}
	return &cpv, nil
}

func (k Keeper) VerifyConsumerValidatorSet(ctx sdk.Context, clientID string, valSet restaking.ValidatorSet) error {
	consensusState, found := k.clientKeeper.GetLatestClientConsensusState(ctx, clientID)
	if !found {
		return fmt.Errorf("client consensus status is not exist %s", clientID)
	}

	tmConsensusState, ok := consensusState.(*ibctmtypes.ConsensusState)
	if !ok {
		return fmt.Errorf("client consensus must be kind of tendermint %s", clientID)
	}

	tmValSet, err := tmtypes.ValidatorSetFromProto(&valSet)
	if err != nil {
		return err
	}

	if !bytes.Equal(tmConsensusState.NextValidatorsHash, tmValSet.Hash()) {
		return fmt.Errorf("validator hash mismatch %s", clientID)
	}

	return nil
}

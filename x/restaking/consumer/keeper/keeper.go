package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	restaking "github.com/celinium-network/celinium/x/restaking/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	cdc          codec.Codec
	scopedKeeper restaking.ScopedKeeper

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper restaking.IBCTransferKeeper
	ibcCoreKeeper     restaking.IBCCoreKeeper

	standaloneStakingKeeper restaking.StakingKeeper
	slashingKeeper          restaking.SlashingKeeper
	bankKeeper              restaking.BankKeeper
	authKeeper              restaking.AccountKeeper
}

func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.Codec,
	scopedKeeper restaking.ScopedKeeper,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper restaking.IBCTransferKeeper,
	ibcCoreKeeper restaking.IBCCoreKeeper,
	standaloneStakingKeeper restaking.StakingKeeper,
	slashingKeeper restaking.SlashingKeeper,
	bankKeeper restaking.BankKeeper,
	authKeeper restaking.AccountKeeper,
) Keeper {
	k := Keeper{
		storeKey:                storeKey,
		cdc:                     cdc,
		scopedKeeper:            scopedKeeper,
		channelKeeper:           channelKeeper,
		portKeeper:              portKeeper,
		connectionKeeper:        connectionKeeper,
		clientKeeper:            clientKeeper,
		ibcTransferKeeper:       ibcTransferKeeper,
		ibcCoreKeeper:           ibcCoreKeeper,
		standaloneStakingKeeper: standaloneStakingKeeper,
		slashingKeeper:          slashingKeeper,
		bankKeeper:              bankKeeper,
		authKeeper:              authKeeper,
	}
	return k
}

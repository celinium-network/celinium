package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddSourceChain{}, "interstaking/MsgAddSourceChain", nil)
	cdc.RegisterConcrete(&MsgDelegate{}, "interstaking/MsgDelegate", nil)
	cdc.RegisterConcrete(&SourceChainMetadata{}, "interstaking/SourceChainMetadata", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgAddSourceChain{},
		&MsgDelegate{},
	)
}
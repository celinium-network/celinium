package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MarshalProtoType(cdc codec.BinaryCodec, t codec.ProtoMarshaler) ([]byte, error) {
	return cdc.Marshal(t)
}

func MustMarshalProtoType(cdc codec.BinaryCodec, t codec.ProtoMarshaler) []byte {
	bz, err := cdc.Marshal(t)
	if err != nil {
		panic(err)
	}
	return bz
}

func UnMarshalProtoType(cdc codec.BinaryCodec, bz []byte, t codec.ProtoMarshaler) error {
	return cdc.UnmarshalInterface(bz, t)
}

func MustUnMarshalProtoType(cdc codec.BinaryCodec, bz []byte, t codec.ProtoMarshaler) {
	err := UnMarshalProtoType(cdc, bz, t)
	if err != nil {
		panic(err)
	}
}

func GenerateSourceChainControlAccount(ctx sdk.Context, chainID string, connectionID string) authtypes.AccountI {
	header := ctx.BlockHeader()

	buf := []byte(ModuleName + chainID + connectionID)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	return authtypes.NewEmptyModuleAccount(string(buf), authtypes.Staking)
}

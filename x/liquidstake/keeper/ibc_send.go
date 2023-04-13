package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	"github.com/gogo/protobuf/proto"
)

func (k Keeper) sendIBCMsg(ctx sdk.Context, msgs []proto.Message, connectionID string, sender string) (uint64, string, error) {
	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return 0, "", err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// TODO timeout ?
	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	sendPortID, err := icatypes.NewControllerPortID(sender)
	if err != nil {
		return 0, "", err
	}

	sequence, err := k.icaCtlKeeper.SendTx(ctx, nil, connectionID, sendPortID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return 0, "", err
	}
	return sequence, sendPortID, nil
}

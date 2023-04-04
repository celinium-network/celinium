package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"celinium/x/liquidstake/types"
)

func (k Keeper) SetCallBack(ctx sdk.Context, channel string, port string, sequence uint64, callback *types.IBCCallback) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(callback)
	store.Set(types.GetIBCDelegationCallbackKey([]byte(channel), []byte(port), sequence), bz)
}

func (k Keeper) GetCallBack(ctx sdk.Context, channel string, port string, sequence uint64) (*types.IBCCallback, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetIBCDelegationCallbackKey([]byte(channel), []byte(port), sequence))
	if bz == nil {
		return nil, false
	}

	callback := types.IBCCallback{}
	k.cdc.MustUnmarshal(bz, &callback)

	return &callback, true
}

func (k Keeper) HandleIBCAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, responses []*codectypes.Any) error {
	callback, found := k.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		return nil
	}

	successful := callback.CheckSuccessfulIBCAcknowledgement(k.cdc, responses)

	// update record status
	k.advanceCallbackRelatedEntry(ctx, callback, successful)

	// TODO consider remove callback ?, repeated receive same Acknowledgement
	return nil
}

func (k Keeper) advanceCallbackRelatedEntry(ctx sdk.Context, callback *types.IBCCallback, successful bool) {
	// TODO optimize the if/else code block
	if callback.CallType == types.DelegateTransferCall {
		delegationRecordID := sdk.BigEndianToUint64([]byte(callback.Args))
		record, found := k.GetDelegationRecord(ctx, delegationRecordID)
		if !found {
			return
		}

		k.AfterDelegateTransfer(ctx, record, successful)

	} else if callback.CallType == types.DelegateCall {
		delegationRecordID := sdk.BigEndianToUint64([]byte(callback.Args))
		record, found := k.GetDelegationRecord(ctx, delegationRecordID)
		if !found {
			return
		}

		k.AfterCrosschainDelegate(ctx, record, successful)
	}
}

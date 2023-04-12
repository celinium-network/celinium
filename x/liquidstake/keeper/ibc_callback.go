package keeper

import (
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
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

func GetResultFromAcknowledgement(acknowledgement []byte) ([]byte, error) {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return nil, err
	}

	switch response := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		if len(response.Result) == 0 {
			return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "empty acknowledgement")
		}
		return ack.GetResult(), nil
	case *channeltypes.Acknowledgement_Error:
		return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidPacket, "invalid acknowledgement")
	default:
		return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "unknown acknowledgement status")

	}
}

func (k Keeper) HandleIBCTransferAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, acknowledgement []byte) error {
	_, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	callback, found := k.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		return sdkerrors.Wrapf(types.ErrCallbackNotExist, "channelID: %s, portID: %s, sequence: %d",
			packet.SourceChannel, packet.SourcePort, packet.Sequence)
	}

	// update record status
	k.advanceCallbackRelatedEntry(ctx, callback, nil, true)

	// TODO consider remove callback ?, repeated receive same Acknowledgement
	return nil
}

func (k Keeper) HandleICAAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, acknowledgement []byte) error {
	res, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var txMsgData sdk.TxMsgData
	if err := k.cdc.Unmarshal(res, &txMsgData); err != nil {
		return err
	}

	callback, found := k.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		return sdkerrors.Wrapf(types.ErrCallbackNotExist, "channelID: %s, portID: %s, sequence: %d",
			packet.SourceChannel, packet.SourcePort, packet.Sequence)
	}

	successful := callback.CheckSuccessfulIBCAcknowledgement(k.cdc, txMsgData.MsgResponses)

	// update record status
	k.advanceCallbackRelatedEntry(ctx, callback, txMsgData.MsgResponses, successful)

	// TODO consider remove callback ?, repeated receive same Acknowledgement
	return nil
}

func (k Keeper) advanceCallbackRelatedEntry(ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any, successful bool) {
	// TODO too much swiach/case
	switch callback.CallType {
	case types.DelegateTransferCall:
		delegationRecordID := sdk.BigEndianToUint64([]byte(callback.Args))
		record, found := k.GetDelegationRecord(ctx, delegationRecordID)
		if !found {
			return
		}

		k.AfterDelegateTransfer(ctx, record, successful)
	case types.DelegateCall:
		delegationRecordID := sdk.BigEndianToUint64([]byte(callback.Args))
		record, found := k.GetDelegationRecord(ctx, delegationRecordID)
		if !found {
			return
		}

		k.AfterCrosschainDelegate(ctx, record, successful)
	case types.UnbondCall:
		var completeTime time.Time
		for _, r := range responses {
			if strings.Contains(r.TypeUrl, "MsgUndelegateResponse") {
				response := stakingtypes.MsgUndelegateResponse{}
				if err := k.cdc.Unmarshal(r.Value, &response); err != nil {
					return
				}
				completeTime = response.CompletionTime
			}
		}
		var unbondCallArgs types.UnbondCallbackArgs

		k.cdc.MustUnmarshal([]byte(callback.Args), &unbondCallArgs)

		epochUnbondings, found := k.GetEpochUnboundings(ctx, unbondCallArgs.Epoch)
		if !found {
			return
		}

		for i := 0; i < len(epochUnbondings.Unbondings); i++ {
			if epochUnbondings.Unbondings[i].ChainID != unbondCallArgs.ChainID {
				continue
			}
			epochUnbondings.Unbondings[i].UnbondTIme = uint64(completeTime.UnixNano())
			epochUnbondings.Unbondings[i].Status = types.UnbondingWaitting
		}

		// save
		k.SetEpochUnboundings(ctx, epochUnbondings)

		// TODO remove SourceChain.StakedAmount
	case types.WithdrawUnbondCall:
		var unbondCallArgs types.UnbondCallbackArgs
		k.cdc.MustUnmarshal([]byte(callback.Args), &unbondCallArgs)
		epochUnbondings, found := k.GetEpochUnboundings(ctx, unbondCallArgs.Epoch)
		if !found {
			return
		}

		for _, unbonding := range epochUnbondings.Unbondings {
			if unbonding.ChainID != unbondCallArgs.ChainID {
				continue
			}
			unbonding.Status = types.UnbondingDone

			for _, userUnDelegationID := range unbonding.UserUnbondRecordIds {
				userUndelegation, found := k.GetUndelegationRecordByID(ctx, userUnDelegationID)
				if !found {
					continue
				}
				userUndelegation.CliamStatus = types.UndelegationClaimable
				k.SetUndelegationRecord(ctx, userUndelegation)
			}
		}
		k.SetEpochUnboundings(ctx, epochUnbondings)
	case types.TransferRewardCall:
		var callbackArgs types.TransferRewardCallbackArgs
		k.cdc.MustUnmarshal([]byte(callback.Args), &callbackArgs)
		epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, types.DelegationEpochIdentifier)
		if !found {
			return
		}

		currentEpoch := uint64(epochInfo.CurrentEpoch)
		recordID, found := k.GetChianDelegationRecordID(ctx, callbackArgs.ChainID, currentEpoch)
		if !found {
			return
		}

		record, found := k.GetDelegationRecord(ctx, recordID)
		if !found {
			return
		}

		record.DelegationCoin = record.DelegationCoin.AddAmount(callbackArgs.Amount)

		k.SetDelegationRecord(ctx, recordID, record)
	case types.SetWithdrawAddressCall:
		// TODO make source chain available here
	default:
	}
}

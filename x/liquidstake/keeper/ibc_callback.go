package keeper

import (
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	k.advanceCallbackRelatedEntry(ctx, callback, responses, successful)

	// TODO consider remove callback ?, repeated receive same Acknowledgement
	return nil
}

func (k Keeper) advanceCallbackRelatedEntry(ctx sdk.Context, callback *types.IBCCallback, responses []*codectypes.Any, successful bool) {
	// TODO optimize the if/else code block
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

		for _, unbonding := range epochUnbondings.Unbondings {
			if unbonding.ChainID != unbondCallArgs.ChainID {
				continue
			}
			unbonding.UnbondTIme = uint64(completeTime.Unix())
			unbonding.Status = types.UnbondingWaitting
		}
		// save
		k.SetEpochUnboundings(ctx, epochUnbondings)
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

	default:
	}
}

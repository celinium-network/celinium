package keeper

import (
	"math"
	"time"

	"celinium/x/inter-staking/types"

	sdkmath "cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	tmlightclienttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	"github.com/gogo/protobuf/proto"
)

func (k Keeper) UnDelegate(ctx sdk.Context, chainID string, undelegationAmount sdk.Coin, delegator string) error {
	sourceChainMetadata, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}
	// check wheather the ica of source chain in control endpoint is available.
	if !k.SourceChainAvaiable(ctx, sourceChainMetadata.IbcConnectionId, sourceChainMetadata.ICAControlAddr) {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	// Check wheather the coin is the native token of the source chain.
	// if strings.Compare(undelegationAmount.Denom, sourceChainMetadata.SourceChainTraceDenom) != 0 {
	// 	return sdkerrors.Wrapf(types.ErrMismatchSourceCoin, "chainID: %s, expected: %s, get:",
	// 		chainID, sourceChainMetadata.SourceChainTraceDenom, undelegationAmount.Denom)
	// }

	delegationAmount := k.GetDelegation(ctx, delegator, chainID)
	if delegationAmount.Amount.LT(undelegationAmount.Amount) {
		return sdkerrors.Wrapf(types.ErrInsufficientDelegation, "exist: %s, expected %s", delegationAmount, undelegationAmount)
	}

	remindingAmount := delegationAmount.Sub(undelegationAmount)
	// remove key if remindingAmount == 0?
	k.SetDelegation(&ctx, delegator, chainID, remindingAmount)

	stragegyLen := len(sourceChainMetadata.DelegateStrategy)

	undelegateMsgs := make([]proto.Message, 0)
	usedAmount := sdkmath.NewInt(0)

	portID, err := icatypes.NewControllerPortID(sourceChainMetadata.ICAControlAddr)
	if err != nil {
		return nil
	}
	hostAddr, ok := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, sourceChainMetadata.IbcConnectionId, portID)
	if !ok {
		return sdk.ErrEmptyHexAddress
	}

	hostAddress := sdk.MustAccAddressFromBech32(hostAddr)

	for i := 0; i < stragegyLen-2; i++ {
		percentage := sdkmath.NewIntFromUint64(uint64(sourceChainMetadata.DelegateStrategy[i].Percentage))
		stakingAmount := undelegationAmount.Amount.Mul(percentage).BigInt()
		stakingAmount.Div(stakingAmount, types.PercentageDenominator.BigInt())
		usedAmount.Add(sdkmath.NewIntFromBigInt(stakingAmount))

		valAddress, err := sdk.ValAddressFromBech32(sourceChainMetadata.DelegateStrategy[i].ValidatorAddress)
		if err != nil {
			return err
		}

		undelegateMsgs = append(undelegateMsgs, stakingtypes.NewMsgUndelegate(
			hostAddress,
			valAddress,
			sdk.NewCoin(sourceChainMetadata.SourceChainDenom, sdkmath.NewIntFromBigInt(stakingAmount)),
		))
	}

	if !usedAmount.Equal(delegationAmount.Amount) {
		valAddress, err := sdk.ValAddressFromBech32(sourceChainMetadata.DelegateStrategy[stragegyLen-1].ValidatorAddress)
		if err != nil {
			return err
		}
		undelegateMsgs = append(undelegateMsgs, stakingtypes.NewMsgUndelegate(
			hostAddress,
			valAddress,
			sdk.NewCoin(sourceChainMetadata.SourceChainDenom, delegationAmount.Amount.Sub(usedAmount)),
		))
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, undelegateMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	sequence, err := k.icaControllerKeeper.SendTx(ctx, nil, sourceChainMetadata.IbcConnectionId, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}

	// push into undelegation pending queue. I think a unique queue is needed ?
	k.PushDelegationTaskQueue(&ctx, types.PendingUndelegationQueueKey, sequence, &types.DelegationTask{
		ChainId:   chainID,
		Delegator: delegator,
		Amount:    undelegationAmount,
	})

	return nil
}

func (k Keeper) OnUndelegateAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, resp *stakingtypes.MsgUndelegateResponse) {
	pendingUndelegation := k.GetDelegationQueueSlice(&ctx, types.PendingUndelegationQueueKey, packet.Sequence)
	if len(pendingUndelegation) == 0 {
		return
	}

	for _, t := range pendingUndelegation {
		k.SetPreparingUndelegation(ctx, resp.CompletionTime, packet.Sequence, t.ChainId, t.Delegator)
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDelegateQueueKey(types.PreparingDelegationQueueKey, packet.Sequence))
}

func (k Keeper) SubmitSourceChainUnbondingDelegation(
	ctx sdk.Context,
	chainID string,
	clientID string,
	proofs [][]byte,
	proofHeight exported.Height,
	unbondingDelegations []stakingtypes.UnbondingDelegation,
) error {
	for i, proof := range proofs {
		if err := k.VerifyUnboundingDelegation(ctx, clientID, proof, proofHeight, &unbondingDelegations[i]); err != nil {
			return err
		}
	}

	sourceChainMetadata, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	if len(unbondingDelegations) != len(sourceChainMetadata.DelegateStrategy) {
		return sdkerrors.Wrapf(
			types.ErrSubmitSourceChainUnbondingQueue,
			"source chain metadata delegate strategy expect len:%s, actual: %s",
			len(sourceChainMetadata.DelegateStrategy),
			len(unbondingDelegations))
	}

	lastUbd := k.GetSourceChainUnbondingDelegations(ctx, chainID, clientID)
	if lastUbd == nil {
		k.SetSourceChainUnbondingDelegations(ctx, types.SourceChainUnbondingQueue{
			ChainID:              chainID,
			ClientID:             clientID,
			LastHeight:           proofHeight.GetRevisionHeight(),
			UnbondingDelegations: unbondingDelegations,
		})
		return nil
	}

	// if proofHeight.GetRevisionHeight()-lastUbd.LastHeight > 5 {
	// 	return types.ErrSubmitTimeOut
	// }

	completeTime := unbondingDelegations[0].Entries[0].CompletionTime
	entryLen := len(lastUbd.UnbondingDelegations[0].Entries)
	delegationLen := len(unbondingDelegations)
	done := false

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingUndelegationQueueKey)

	for i := 0; i < entryLen && !done; i++ {
		var removeAmount sdkmath.Int
		entryCompleteTime := lastUbd.UnbondingDelegations[0].Entries[i].CompletionTime
		if entryCompleteTime.Before(completeTime) || entryCompleteTime.Equal(completeTime) {
			break
		}
		for j := 0; j < delegationLen; j++ {
			removeAmount.Add(lastUbd.UnbondingDelegations[0].Entries[i].Balance)
		}

		if !iterator.Valid() {
			break
		}

		bz := iterator.Value()
		undelegation := types.DelegationTask{}
		k.cdc.MustUnmarshal(bz, &undelegation)

		// delete key after cross chain transfer successfully
		store.Delete(iterator.Key())

		// assemble message into array, send once ??
		k.Distribute(
			ctx,
			chainID,
			sdk.Coin{Denom: sourceChainMetadata.SourceChainDenom, Amount: removeAmount},
			sourceChainMetadata,
			undelegation.Delegator)

		iterator.Next()
	}

	return nil
}

func (k *Keeper) SubmitSourceChainDVPairNotExist(
	ctx sdk.Context,
	chainID string,
	clientID string,
	proofs [][]byte,
	proofHeight exported.Height,
	dvPairs []stakingtypes.DVPair,
) error {
	for i, proof := range proofs {
		delAddr := sdk.MustAccAddressFromBech32(dvPairs[i].DelegatorAddress)
		valAddr, err := sdk.ValAddressFromBech32(dvPairs[i].ValidatorAddress)
		if err != nil {
			panic(err)
		}
		key := stakingtypes.GetUBDKey(delAddr, valAddr)
		if err := k.VerifyUnboundingDelegationNoExist(ctx, clientID, proof, proofHeight, string(key)); err != nil {
			return err
		}
	}

	lastUbd := k.GetSourceChainUnbondingDelegations(ctx, chainID, clientID)

	if lastUbd == nil {
		// maybe returen err?
		return nil
	}

	sourceChainMetadata, found := k.GetSourceChain(ctx, chainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownSourceChain, "chainID: %s", chainID)
	}

	entryLen := len(lastUbd.UnbondingDelegations[0].Entries)
	delegationLen := len(lastUbd.UnbondingDelegations)
	done := false

	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingUndelegationQueueKey)

	for i := 0; i < entryLen && !done; i++ {
		removeAmount := sdkmath.NewIntFromUint64(0)
		for j := 0; j < delegationLen; j++ {
			removeAmount = removeAmount.Add(lastUbd.UnbondingDelegations[0].Entries[i].Balance)
		}

		if !iterator.Valid() {
			break
		}

		bz := iterator.Value()
		undelegations := types.DelegationTasks{}
		k.cdc.MustUnmarshal(bz, &undelegations)

		// delete key after cross chain transfer successfully
		store.Delete(iterator.Key())

		// assemble message into array, send once ??
		err := k.Distribute(
			ctx,
			chainID,
			sdk.Coin{Denom: sourceChainMetadata.SourceChainDenom, Amount: removeAmount},
			sourceChainMetadata,
			undelegations.DelegationTasks[0].Delegator)
		if err != nil {
			return err
		}

		iterator.Next()
	}

	return nil
}

func (k Keeper) VerifyUnboundingDelegation(
	ctx sdk.Context,
	clientID string,
	proof []byte,
	proofHeight exported.Height,
	unbondDelegation *stakingtypes.UnbondingDelegation,
) error {
	clientStore := k.ibcClientKeeper.ClientStore(ctx, clientID)

	targetClient, found := k.ibcClientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	// if status := targetClient.Status(ctx, clientStore, k.cdc); status != exported.Active {
	// 	return sdkerrors.Wrapf(clienttypes.ErrClientNotActive, "client (%s) status is %s", clientID, status)
	// }

	if targetClient.GetLatestHeight().LT(proofHeight) {
		return sdk.ErrIntOverflowAbci
	}

	tmTargetClient, ok := targetClient.(*tmlightclienttypes.ClientState)
	if !ok {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	var merkleProof commitmenttypes.MerkleProof

	if err := k.cdc.Unmarshal(proof, &merkleProof); err != nil {
		return err
	}

	consensusState, err := tmlightclienttypes.GetConsensusState(clientStore, k.cdc, proofHeight)
	if err != nil {
		sdkerrors.Wrap(err, "please ensure the proof was constructed against a height that exists on the client")
	}

	bz, err := k.cdc.Marshal(unbondDelegation)
	if err != nil {
		return err
	}

	delAddr := sdk.MustAccAddressFromBech32(unbondDelegation.DelegatorAddress)
	valAddr, err := sdk.ValAddressFromBech32(unbondDelegation.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	key := stakingtypes.GetUBDKey(delAddr, valAddr)

	path := commitmenttypes.NewMerklePath(append([]string{stakingtypes.StoreKey}, string(key))...)

	if err = merkleProof.VerifyMembership(tmTargetClient.ProofSpecs, consensusState.GetRoot(), path, bz); err != nil {
		return err
	}

	return nil
}

func (k Keeper) Distribute(ctx sdk.Context, chainID string, amount sdk.Coin, metadata *types.SourceChainMetadata, receiver string) error {
	portID, _ := icatypes.NewControllerPortID(metadata.ICAControlAddr)

	hostAddr, _ := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, metadata.IbcConnectionId, portID)

	timeoutHeight := clienttypes.NewHeight(math.MaxUint64, math.MaxUint64)
	transferMsg := transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		metadata.IbcTransferChannelId,
		amount,
		hostAddr,
		receiver,
		timeoutHeight,
		0,
		"",
	)

	protoMsg := make([]proto.Message, 0)
	protoMsg = append(protoMsg, transferMsg)
	data, err := icatypes.SerializeCosmosTx(k.cdc, protoMsg)
	if err != nil {
		return nil
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	_, err = k.icaControllerKeeper.SendTx(ctx, nil, metadata.IbcConnectionId, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) VerifyUnboundingDelegationNoExist(
	ctx sdk.Context,
	clientID string,
	proof []byte,
	proofHeight exported.Height,
	key string,
) error {
	clientStore := k.ibcClientKeeper.ClientStore(ctx, clientID)

	targetClient, found := k.ibcClientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	// if status := targetClient.Status(ctx, clientStore, k.cdc); status != exported.Active {
	// 	return sdkerrors.Wrapf(clienttypes.ErrClientNotActive, "client (%s) status is %s", clientID, status)
	// }

	if targetClient.GetLatestHeight().LT(proofHeight) {
		return sdk.ErrIntOverflowAbci
	}

	tmTargetClient, ok := targetClient.(*tmlightclienttypes.ClientState)
	if !ok {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	var merkleProof commitmenttypes.MerkleProof

	if err := k.cdc.Unmarshal(proof, &merkleProof); err != nil {
		return err
	}

	consensusState, err := tmlightclienttypes.GetConsensusState(clientStore, k.cdc, proofHeight)
	if err != nil {
		return sdkerrors.Wrap(err, "please ensure the proof was constructed against a height that exists on the client")
	}

	path := commitmenttypes.NewMerklePath(append([]string{stakingtypes.StoreKey}, string(key))...)

	err = merkleProof.VerifyNonMembership(tmTargetClient.ProofSpecs, consensusState.GetRoot(), path)

	return err
}

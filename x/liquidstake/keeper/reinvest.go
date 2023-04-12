package keeper

import (
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	tmlightclienttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/gogo/protobuf/proto"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (k Keeper) StartInvest(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SouceChainKeyPrefix)

	for ; iterator.Valid(); iterator.Next() {
		sourcechain := &types.SourceChain{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, sourcechain)

		if !k.sourceChainAvaiable(ctx, sourcechain) {
			continue
		}

		k.QueryReward(ctx, sourcechain)
	}
}

func (k Keeper) SubmitQueryRewardReponse(
	ctx sdk.Context,
	query *types.IBCQuery,
	queryHeight uint64,
	proof []byte,
	proofHeight exported.Height,
	response []byte,
) error {
	paths := strings.Split(query.QueryPathKey, "-")
	k.verifyMerkleProof(ctx, query.ConnectionID, proof, proofHeight, paths, response)

	queryID := query.ID(queryHeight)

	_, found := k.GetIBCQuery(ctx, []byte(queryID))

	if !found {
		return sdkerrors.Wrapf(types.ErrIBCQueryNotExist, "query %v", query)
	}

	var rewardAmount math.Int
	rewardAmount.Unmarshal(response)

	sendMsgs := make([]proto.Message, 0)

	sourceChain, found := k.GetSourceChain(ctx, query.ChainID)
	if !found {
		return sdkerrors.Wrapf(types.ErrSourceChainExist, "chainID: %s", query.ChainID)
	}

	delegateAccAddr, err := k.getSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.UnboudAddress)
	if err != nil {
		return err
	}

	rewardAccAddr, err := k.getSourceChainAddr(ctx, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
	if err != nil {
		return err
	}

	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: rewardAccAddr.String(),
		ToAddress:   delegateAccAddr.String(),
		Amount:      []sdk.Coin{},
	})

	data, err := icatypes.SerializeCosmosTx(k.cdc, sendMsgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// TODO timeout ?
	timeoutTimestamp := ctx.BlockTime().Add(30 * time.Minute).UnixNano()
	portID, err := icatypes.NewControllerPortID(sourceChain.WithdrawAddress)
	if err != nil {
		return err
	}

	sequence, err := k.icaCtlKeeper.SendTx(ctx, nil, sourceChain.ConnectionID, portID, packetData, uint64(timeoutTimestamp)) //nolint:staticcheck //
	if err != nil {
		return err
	}

	callbackArgs := types.TransferRewardCallbackArgs{
		ChainID: sourceChain.ChainID,
		Amount:  rewardAmount,
	}

	callbackArgsBz := k.cdc.MustMarshal(&callbackArgs)

	callback := types.IBCCallback{
		CallType: types.TransferRewardCall,
		Args:     string(callbackArgsBz),
	}

	sendChannelID, _ := k.icaCtlKeeper.GetOpenActiveChannel(ctx, sourceChain.ConnectionID, portID)

	k.SetCallBack(ctx, sendChannelID, portID, sequence, &callback)

	return nil
}

func (k Keeper) getSourceChainAddr(ctx sdk.Context, connectionID string, ctlAddress string) (sdk.AccAddress, error) {
	portID, err := icatypes.NewControllerPortID(ctlAddress)
	if err != nil {
		return nil, err
	}

	sourceChainDelegateAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, connectionID, portID)
	return sdk.AccAddressFromBech32(sourceChainDelegateAddr)
}

func (k Keeper) verifyMerkleProof(
	ctx sdk.Context,
	connectionID string,
	proof []byte,
	proofHeight exported.Height,
	key []string,
	value []byte,
) error {
	connection, found := k.ibcKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return nil
	}
	targetClient, found := k.ibcKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.ClientId)
	}

	var merkleProof commitmenttypes.MerkleProof

	if err := k.cdc.Unmarshal(proof, &merkleProof); err != nil {
		return err
	}

	tmTargetClient, ok := targetClient.(*tmlightclienttypes.ClientState)
	if !ok {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.ClientId)
	}

	clientStore := k.ibcKeeper.ClientKeeper.ClientStore(ctx, connection.ClientId)

	consensusState, err := tmlightclienttypes.GetConsensusState(clientStore, k.cdc, proofHeight)
	if err != nil {
		return sdkerrors.Wrap(err, "please ensure the proof was constructed against a height that exists on the client")
	}

	merklepath := commitmenttypes.NewMerklePath(key...)

	if err = merkleProof.VerifyMembership(tmTargetClient.ProofSpecs, consensusState.GetRoot(), merklepath, value); err != nil {
		return err
	}
	return nil
}

func (k Keeper) QueryReward(ctx sdk.Context, sourceChain *types.SourceChain) error {
	portID, err := icatypes.NewControllerPortID(sourceChain.WithdrawAddress)
	if err != nil {
		return err
	}

	withdrawHostAddr, _ := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, portID)

	_, withdrawHostAddrBz, err := bech32.DecodeAndConvert(withdrawHostAddr)
	if err != nil {
		return err
	}

	queryBz := append(banktypes.CreateAccountBalancesPrefix(withdrawHostAddrBz), []byte(sourceChain.NativeDenom)...)

	epochInfo, found := k.epochKeeper.GetEpochInfo(ctx, types.DelegationEpochIdentifier)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownEpoch, "unknown epoch %s", types.DelegationEpochIdentifier)
	}

	timeout := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration / 2).UnixNano()

	balanceMerkleKeyPath := append([]string{banktypes.StoreKey}, string(queryBz))
	queryPathKey := strings.Join(balanceMerkleKeyPath, "-")

	query := types.IBCQuery{
		QueryType:    "balance",
		QueryPathKey: queryPathKey,
		Timeout:      uint64(timeout),
		ChainID:      sourceChain.ChainID,
		ConnectionID: sourceChain.ConnectionID,
	}

	k.SetIBCQuery(ctx, &query)

	return nil
}

func (k Keeper) SetIBCQuery(ctx sdk.Context, query *types.IBCQuery) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(query)

	store.Set([]byte(query.ID(uint64(ctx.BlockHeight()))), bz)
}

func (k Keeper) GetIBCQuery(ctx sdk.Context, id []byte) (*types.IBCQuery, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(id)
	if bz == nil {
		return nil, false
	}

	query := types.IBCQuery{}
	k.cdc.MustUnmarshal(bz, &query)

	return &query, true
}

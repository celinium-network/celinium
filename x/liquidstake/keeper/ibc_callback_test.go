package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	liquidstaketypes "github.com/celinium-network/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestHandleDelegateTransferIBCAck() {
	srcChainParams := suite.mockSourceChainParams()
	epoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, epoch)

	ctx := suite.controlChain.GetContext()
	ctlChainApp := getCeliniumApp(suite.controlChain)

	testCases := []struct {
		msg        string
		delegation liquidstaketypes.ProxyDelegation
		callback   liquidstaketypes.IBCCallback
		ack        channeltypes.Acknowledgement
		packet     channeltypes.Packet
		checker    func(*channeltypes.Packet, *liquidstaketypes.ProxyDelegation)
	}{
		{
			"bad ack", // TODO check ibc timeout will get `ErrorAcknowledgement` ?
			liquidstaketypes.ProxyDelegation{
				Id:             1,
				Status:         liquidstaketypes.ProxyDelegationTransferring,
				EpochNumber:    uint64(epoch.CurrentEpoch),
				ChainID:        srcChainParams.ChainID,
				ReinvestAmount: math.Int{},
			},
			liquidstaketypes.IBCCallback{
				CallType: liquidstaketypes.DelegateTransferCall,
				Args:     string(sdk.Uint64ToBigEndian(1)),
			},
			channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed")),
			channeltypes.Packet{
				Sequence:      0,
				SourcePort:    "transfer",
				SourceChannel: "channel-0",
			},
			func(packet *channeltypes.Packet, pd *liquidstaketypes.ProxyDelegation) {
				newDelegation, found := ctlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, pd.Id)
				suite.Require().True(found)
				suite.Require().Equal(newDelegation.Status, liquidstaketypes.ProxyDelegationPending)

				_, found = ctlChainApp.LiquidStakeKeeper.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
				suite.Require().True(found)
			},
		},
		{
			"right ack",
			liquidstaketypes.ProxyDelegation{
				Id:             2,
				Status:         liquidstaketypes.ProxyDelegationTransferring,
				EpochNumber:    uint64(epoch.CurrentEpoch),
				ChainID:        srcChainParams.ChainID,
				ReinvestAmount: math.Int{},
			},
			liquidstaketypes.IBCCallback{
				CallType: liquidstaketypes.DelegateTransferCall,
				Args:     string(sdk.Uint64ToBigEndian(2)),
			},
			channeltypes.NewResultAcknowledgement([]byte("successful")),
			channeltypes.Packet{
				Sequence:      2,
				SourcePort:    "transfer",
				SourceChannel: "channel-0",
			},
			func(packet *channeltypes.Packet, pd *liquidstaketypes.ProxyDelegation) {
				newDelegation, found := ctlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, pd.Id)
				suite.Require().True(found)
				suite.Require().Equal(newDelegation.Status, liquidstaketypes.ProxyDelegating)

				_, found = ctlChainApp.LiquidStakeKeeper.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
				suite.Require().False(found)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			ctlChainApp.LiquidStakeKeeper.SetProxyDelegation(ctx, tc.delegation.Id, &tc.delegation)
			ctlChainApp.LiquidStakeKeeper.SetCallBack(ctx, tc.packet.SourceChannel, tc.packet.SourcePort, tc.packet.Sequence, &tc.callback)

			ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&tc.ack)
			ctlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(ctx, &tc.packet, ackBz)
			tc.checker(&tc.packet, &tc.delegation)
		})
	}
}

func (suite *KeeperTestSuite) TestHandleDelegateIBC_WithErrorAck() {
	epoch := suite.delegationEpoch()
	srcChainParams := suite.mockSourceChainParams()
	suite.setSourceChainAndEpoch(srcChainParams, epoch)

	cdc := suite.controlChain.Codec
	ctx := suite.controlChain.GetContext()
	ctlChainApp := getCeliniumApp(suite.controlChain)

	delegation := liquidstaketypes.ProxyDelegation{
		Id:             1,
		Status:         liquidstaketypes.ProxyDelegationTransferred,
		EpochNumber:    uint64(epoch.CurrentEpoch),
		ChainID:        srcChainParams.ChainID,
		ReinvestAmount: math.Int{},
	}

	srcChainParams.AllocateTokenForValidator(sdk.NewIntFromUint64(10000))

	callback := liquidstaketypes.IBCCallback{
		CallType: liquidstaketypes.DelegateCall,
		Args: string(cdc.MustMarshal(&liquidstaketypes.DelegateCallbackArgs{
			Validators:        srcChainParams.Validators,
			ProxyDelegationID: 1,
		})),
	}

	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed"))
	packet := channeltypes.Packet{
		Sequence:      0,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	ctlChainApp.LiquidStakeKeeper.SetProxyDelegation(ctx, delegation.Id, &delegation)
	ctlChainApp.LiquidStakeKeeper.SetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence, &callback)

	ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&ack)

	ctlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(ctx, &packet, ackBz)

	_, found := ctlChainApp.LiquidStakeKeeper.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	suite.Require().True(found)

	handledDelegation, _ := ctlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, delegation.Id)
	suite.Require().Equal(handledDelegation.Status, liquidstaketypes.ProxyDelegationFailed)
}

func (suite *KeeperTestSuite) TestHandleDelegateIBC_WithMismatchRespAck() {
	srcChainParams := suite.mockSourceChainParams()
	epoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, epoch)

	ctx := suite.controlChain.GetContext()
	controlChainApp := getCeliniumApp(suite.controlChain)
	cdc := suite.controlChain.Codec

	delegation := liquidstaketypes.ProxyDelegation{
		Id:             1,
		Status:         liquidstaketypes.ProxyDelegationTransferred,
		EpochNumber:    uint64(epoch.CurrentEpoch),
		ChainID:        srcChainParams.ChainID,
		ReinvestAmount: math.Int{},
	}

	srcChainParams.AllocateTokenForValidator(sdk.NewIntFromUint64(10000))

	callback := liquidstaketypes.IBCCallback{
		CallType: liquidstaketypes.DelegateCall,
		Args: string(cdc.MustMarshal(&liquidstaketypes.DelegateCallbackArgs{
			Validators:        srcChainParams.Validators,
			ProxyDelegationID: 1,
		})),
	}

	delegateTxMsg := sdk.TxMsgData{
		MsgResponses: []*codectypes.Any{},
	}

	delegateTxMsgBz := cdc.MustMarshal(&delegateTxMsg)
	ack := channeltypes.NewResultAcknowledgement(delegateTxMsgBz)

	mockPacket := channeltypes.Packet{
		Sequence:      0,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	controlChainApp.LiquidStakeKeeper.SetProxyDelegation(ctx, delegation.Id, &delegation)
	controlChainApp.LiquidStakeKeeper.SetCallBack(ctx, mockPacket.SourceChannel, mockPacket.SourcePort, mockPacket.Sequence, &callback)

	ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&ack)
	controlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(ctx, &mockPacket, ackBz)

	_, found := controlChainApp.LiquidStakeKeeper.GetCallBack(ctx, mockPacket.SourceChannel, mockPacket.SourcePort, mockPacket.Sequence)
	suite.Require().True(found)
	handledDelegation, _ := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, delegation.Id)
	suite.Require().Equal(handledDelegation.Status, liquidstaketypes.ProxyDelegationFailed)
}

func (suite *KeeperTestSuite) TestHandleDelegateIBC_WithCorrectRespAck() {
	srcChainParams := suite.mockSourceChainParams()
	epoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, epoch)

	ctx := suite.controlChain.GetContext()
	controlChainApp := getCeliniumApp(suite.controlChain)
	cdc := suite.controlChain.Codec

	delegation := liquidstaketypes.ProxyDelegation{
		Id:             1,
		Status:         liquidstaketypes.ProxyDelegationTransferred,
		EpochNumber:    uint64(epoch.CurrentEpoch),
		ChainID:        srcChainParams.ChainID,
		ReinvestAmount: math.Int{},
	}

	srcChainParams.AllocateTokenForValidator(sdk.NewIntFromUint64(10000))

	callback := liquidstaketypes.IBCCallback{
		CallType: liquidstaketypes.DelegateCall,
		Args: string(cdc.MustMarshal(&liquidstaketypes.DelegateCallbackArgs{
			Validators:        srcChainParams.Validators,
			ProxyDelegationID: 1,
		})),
	}

	delegateRespMsgs := &stakingtypes.MsgDelegateResponse{}

	delegateRespMsgsVal, err := codectypes.NewAnyWithValue(delegateRespMsgs)
	suite.NoError(err)
	msgResps := make([]*codectypes.Any, len(srcChainParams.Validators))
	for i := 0; i < len(srcChainParams.Validators); i++ {
		msgResps[i] = delegateRespMsgsVal
	}
	delegateTxMsg := sdk.TxMsgData{
		MsgResponses: msgResps,
	}

	delegateTxMsgBz := cdc.MustMarshal(&delegateTxMsg)
	ack := channeltypes.NewResultAcknowledgement(delegateTxMsgBz)

	mockPacket := channeltypes.Packet{
		Sequence:      0,
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
	}
	controlChainApp.LiquidStakeKeeper.SetProxyDelegation(ctx, delegation.Id, &delegation)
	controlChainApp.LiquidStakeKeeper.SetCallBack(ctx, mockPacket.SourceChannel, mockPacket.SourcePort, mockPacket.Sequence, &callback)

	ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&ack)
	controlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(ctx, &mockPacket, ackBz)

	_, found := controlChainApp.LiquidStakeKeeper.GetCallBack(ctx, mockPacket.SourceChannel, mockPacket.SourcePort, mockPacket.Sequence)
	suite.Require().False(found)
	handledDelegation, _ := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, delegation.Id)
	suite.Require().Equal(handledDelegation.Status, liquidstaketypes.ProxyDelegationDone)
}

func (suite *KeeperTestSuite) TestHandleUndelegateIBCAck() {
	var env *mockEpochProxyUnbondingEnv
	var ack channeltypes.Acknowledgement
	var expectedStakedAmount math.Int

	testCases := []struct {
		msg                     string
		malleate                func()
		checker                 func()
		callbackRemoved         bool
		expectedUnbondingStatus liquidstaketypes.ProxyUnbondingStatus
	}{
		{
			msg: "error ack",
			malleate: func() {
				env = suite.mockEpochProxyUnbondingStartedEnv()
				ack = channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed"))
				expectedStakedAmount = env.srcChainParams.StakedAmount
			},
			callbackRemoved:         false,
			expectedUnbondingStatus: liquidstaketypes.ProxyUnbondingStart,
		},
		{
			msg: "mistach ack",
			malleate: func() {
				env = suite.mockEpochProxyUnbondingStartedEnv()
				undelegateTxMsg := sdk.TxMsgData{
					MsgResponses: []*codectypes.Any{},
				}

				delegateTxMsgBz := env.cdc.MustMarshal(&undelegateTxMsg)
				ack = channeltypes.NewResultAcknowledgement(delegateTxMsgBz)
				expectedStakedAmount = env.srcChainParams.StakedAmount
			},
			callbackRemoved:         false,
			expectedUnbondingStatus: liquidstaketypes.ProxyUnbondingStart,
		},
		{
			msg: "correct ack",
			malleate: func() {
				env = suite.mockEpochProxyUnbondingStartedEnv()

				undelegateRespMsgs := &stakingtypes.MsgUndelegateResponse{
					CompletionTime: time.Now().Add(time.Hour * 24),
				}

				undelegateRespMsgsVal, err := codectypes.NewAnyWithValue(undelegateRespMsgs)
				suite.NoError(err)
				msgResps := make([]*codectypes.Any, len(env.srcChainParams.Validators))
				for i := 0; i < len(env.srcChainParams.Validators); i++ {
					msgResps[i] = undelegateRespMsgsVal
				}
				undelegateTxMsg := sdk.TxMsgData{
					MsgResponses: msgResps,
				}

				delegateTxMsgBz := env.cdc.MustMarshal(&undelegateTxMsg)
				ack = channeltypes.NewResultAcknowledgement(delegateTxMsgBz)
				expectedStakedAmount = sdk.ZeroInt()
			},
			callbackRemoved:         true,
			expectedUnbondingStatus: liquidstaketypes.ProxyUnbondingWaitting,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&ack)
			env.ctlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(env.ctx, &env.sendedPacket, ackBz)

			_, foundHandledCallback := env.ctlChainApp.LiquidStakeKeeper.GetCallBack(env.ctx,
				env.sendedPacket.SourceChannel,
				env.sendedPacket.SourcePort,
				env.sendedPacket.Sequence)
			handledProxyUnbonding, _ := env.ctlChainApp.LiquidStakeKeeper.GetEpochProxyUnboundings(env.ctx, env.epoch)
			handledSrcChain, _ := env.ctlChainApp.LiquidStakeKeeper.GetSourceChain(env.ctx, env.srcChainParams.ChainID)

			suite.Require().NotEqual(foundHandledCallback, tc.callbackRemoved)
			suite.Require().Equal(tc.expectedUnbondingStatus, handledProxyUnbonding.Unbondings[0].Status)
			suite.Require().True(handledSrcChain.StakedAmount.Equal(expectedStakedAmount))
		})
	}
}

// func (suite *KeeperTestSuite) TestHandleWithdrawUnbondIBCAck() {
// }

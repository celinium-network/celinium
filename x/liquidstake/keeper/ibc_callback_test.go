package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	liquidstaketypes "github.com/celinium-network/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestHandleDelegateTransferIBCAck() {
	srcChainParams := suite.mockSourceChainParams()
	epoch := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, epoch)

	ctx := suite.controlChain.GetContext()
	controlChainApp := getCeliniumApp(suite.controlChain)

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
				newDelegation, found := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, pd.Id)
				suite.Require().True(found)
				suite.Require().Equal(newDelegation.Status, liquidstaketypes.ProxyDelegationPending)

				_, found = controlChainApp.LiquidStakeKeeper.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
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
				newDelegation, found := controlChainApp.LiquidStakeKeeper.GetProxyDelegation(ctx, pd.Id)
				suite.Require().True(found)
				suite.Require().Equal(newDelegation.Status, liquidstaketypes.ProxyDelegating)

				_, found = controlChainApp.LiquidStakeKeeper.GetCallBack(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
				suite.Require().False(found)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			controlChainApp.LiquidStakeKeeper.SetProxyDelegation(ctx, tc.delegation.Id, &tc.delegation)
			controlChainApp.LiquidStakeKeeper.SetCallBack(ctx, tc.packet.SourceChannel, tc.packet.SourcePort, tc.packet.Sequence, &tc.callback)

			ackBz := channeltypes.SubModuleCdc.MustMarshalJSON(&tc.ack)
			controlChainApp.LiquidStakeKeeper.HandleIBCAcknowledgement(ctx, &tc.packet, ackBz)
			tc.checker(&tc.packet, &tc.delegation)
		})
	}
}

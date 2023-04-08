package keeper_test

import (
	"time"

	"github.com/celinium-netwok/celinium/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

func (suite *KeeperTestSuite) TestIBCTransfer() {
	sourceChainUserAddr := suite.sourceChain.SenderAccount.GetAddress()
	controlChainUserAddr := suite.controlChain.SenderAccount.GetAddress()

	ctlChainApp := getCeliniumApp(suite.controlChain)

	coin := sdk.NewCoin(params.DefaultBondDenom, sdk.NewIntFromUint64(1000000))
	mintCoin(suite.sourceChain, sourceChainUserAddr, coin)

	ibcDenom := suite.calcuateIBCDenom(
		suite.transferPath.EndpointB.ChannelConfig.PortID,
		suite.transferPath.EndpointB.ChannelID,
		params.DefaultBondDenom)

	ctlBalanceBefore := ctlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, ibcDenom)
	suite.IBCTransfer(sourceChainUserAddr.String(), controlChainUserAddr.String(), coin, suite.transferPath, true)

	ctlBalanceAfter := ctlChainApp.BankKeeper.GetBalance(suite.controlChain.GetContext(), controlChainUserAddr, ibcDenom)

	suite.Equal(ctlBalanceAfter.Amount.Sub(ctlBalanceBefore.Amount), coin.Amount)
}

func (suite *KeeperTestSuite) IBCTransfer(
	from string,
	to string,
	coin sdk.Coin,
	transferpath *ibctesting.Path,
	transferForward bool,
) {
	srcEndpoint := transferpath.EndpointA
	destEndpoint := transferpath.EndpointB
	if !transferForward {
		srcEndpoint, destEndpoint = destEndpoint, srcEndpoint
	}

	destChainApp := getCeliniumApp(destEndpoint.Chain)

	timesout := srcEndpoint.Chain.CurrentHeader.Time.Add(time.Hour * 5).UnixNano()
	msg := ibctransfertypes.NewMsgTransfer(
		srcEndpoint.ChannelConfig.PortID,
		srcEndpoint.ChannelID,
		coin,
		from,
		to,
		ibcclienttypes.Height{},
		uint64(timesout),
		"",
	)

	res, err := srcEndpoint.Chain.SendMsgs(msg)
	suite.NoError(err)

	suite.Require().NoError(err)
	suite.transferPath.EndpointB.UpdateClient()

	for _, ev := range res.Events {
		events := sdk.Events{sdk.Event{
			Type:       ev.Type,
			Attributes: ev.Attributes,
		}}

		packet, err := ibctesting.ParsePacketFromEvents(events)
		if err != nil {
			continue
		}

		commitKey := ibchost.PacketCommitmentKey(packet.SourcePort, packet.SourceChannel, packet.Sequence)
		proof, height := srcEndpoint.Chain.QueryProof(commitKey)

		msgRecvPacket := channeltypes.MsgRecvPacket{
			Packet:          packet,
			ProofCommitment: proof,
			ProofHeight:     height,
			Signer:          from,
		}

		_, err = destChainApp.IBCKeeper.RecvPacket(destEndpoint.Chain.GetContext(), &msgRecvPacket)
		suite.NoError(err)
	}
}

func mintCoin(
	chain *ibctesting.TestChain,
	to sdk.AccAddress,
	coin sdk.Coin,
) {
	chainApp := getCeliniumApp(chain)

	chainApp.BankKeeper.MintCoins(
		chain.GetContext(),
		ibctransfertypes.ModuleName,
		sdk.NewCoins(coin),
	)
	chainApp.BankKeeper.SendCoinsFromModuleToAccount(
		chain.GetContext(),
		ibctransfertypes.ModuleName,
		to,
		sdk.NewCoins(coin))
}

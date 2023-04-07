package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v6/testing"

	celiniumapp "github.com/celinium-netwok/celinium/app"
	params "github.com/celinium-netwok/celinium/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestAddSourceChain() {
	sourceChain := types.SourceChain{
		ChainID:                   suite.sourceChain.ChainID,
		ConnectionID:              suite.icaPath.EndpointB.ConnectionID,
		TransferChannelID:         suite.transferPath.EndpointB.ChannelID,
		Bech32ValidatorAddrPrefix: params.Bech32PrefixValAddr,
		Validators: []*types.Validator{
			{
				Address: sdk.ValAddress(suite.sourceChain.Vals.Validators[0].Address).String(),
				Weight:  10000,
			},
		},
		IbcDenom:        suite.calcuateDenomTrace(suite.transferPath.EndpointB),
		NativeDenom:     params.DefaultBondDenom,
		DerivativeDenom: "DLST",
	}

	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	channelSequence := controlChainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := controlChainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), &sourceChain)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	sourceChainAddresses := []string{sourceChain.WithdrawAddress, sourceChain.DelegateAddress, sourceChain.UnboudAddress}

	for _, ica := range sourceChainAddresses {
		portID, err := icatypes.NewControllerPortID(ica)
		suite.NoError(err)
		suite.icaPath.EndpointB.ChannelConfig.PortID = portID
		suite.icaPath.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
		suite.icaPath.EndpointA.ChannelID = suite.icaPath.EndpointB.ChannelID

		channel, _ := controlChainApp.IBCKeeper.ChannelKeeper.GetChannel(suite.controlChain.GetContext(), portID, suite.icaPath.EndpointB.ChannelID)
		suite.icaPath.EndpointB.ChannelConfig.Version = channel.Version

		err = suite.icaPath.EndpointA.ChanOpenTry()
		suite.NoError(err)

		err = suite.icaPath.EndpointB.ChanOpenAck()
		suite.NoError(err)

		err = suite.icaPath.EndpointA.ChanOpenConfirm()
		suite.NoError(err)

		icaFromCtlChain, found := controlChainApp.ICAControllerKeeper.GetInterchainAccountAddress(
			suite.controlChain.GetContext(),
			suite.icaPath.EndpointB.ConnectionID,
			portID)

		suite.True(found)

		icaFromSrcChain, found := sourceChainApp.ICAHostKeeper.GetInterchainAccountAddress(
			suite.sourceChain.GetContext(),
			suite.icaPath.EndpointA.ConnectionID,
			portID)

		suite.True(found)
		suite.Equal(icaFromCtlChain, icaFromSrcChain)

		channelSequence++
	}
}

func getCeliniumApp(chain *ibctesting.TestChain) *celiniumapp.App {
	app, ok := chain.App.(*celiniumapp.App)
	if !ok {
		panic("not celinium app")
	}

	return app
}

func (suite *KeeperTestSuite) calcuateDenomTrace(endpoint *ibctesting.Endpoint) string {
	sourcePrefix := transfertypes.GetDenomPrefix(endpoint.ChannelConfig.PortID, endpoint.ChannelID)
	denomTrace := transfertypes.ParseDenomTrace(sourcePrefix + params.DefaultBondDenom)

	return denomTrace.Hash().String()
}

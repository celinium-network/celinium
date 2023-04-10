package keeper_test

import (
	"math/rand"

	params "github.com/celinium-netwok/celinium/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestAddSourceChain() {
	sourceChain := suite.mockSourceChainParams()

	controlChainApp := getCeliniumApp(suite.controlChain)
	channelSequence := controlChainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := controlChainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), sourceChain)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	createdICAs := getCreatedICAFromSourceChain(sourceChain)

	for _, ica := range createdICAs {
		suite.relayICACreatedPacket(channelSequence, ica)
		channelSequence++
	}
}

func (suite *KeeperTestSuite) relayICACreatedPacket(packetSequence uint64, ctlAccount string) {
	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	portID, err := icatypes.NewControllerPortID(ctlAccount)
	suite.NoError(err)
	suite.icaPath.EndpointB.ChannelConfig.PortID = portID
	suite.icaPath.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(packetSequence)
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
}

func getCreatedICAFromSourceChain(s *types.SourceChain) []string {
	return []string{s.WithdrawAddress, s.DelegateAddress, s.UnboudAddress}
}

func (suite *KeeperTestSuite) mockSourceChainParams() *types.SourceChain {
	sourceChainVals := len(suite.sourceChain.Vals.Validators)
	randVals := rand.Int()%sourceChainVals + 1 //nolint:gosec
	selectedVals := make([]types.Validator, 0)
	maxWeight := uint64(100000)

	for i := 0; i < randVals; i++ {
		selectedVals = append(selectedVals, types.Validator{
			Address: sdk.ValAddress(suite.sourceChain.Vals.Validators[i].Address).String(),
			Weight:  rand.Uint64() % maxWeight, //nolint:gosec
		})
	}

	sourceChain := types.SourceChain{
		ChainID:                   suite.sourceChain.ChainID,
		ConnectionID:              suite.icaPath.EndpointB.ConnectionID,
		TransferChannelID:         suite.transferPath.EndpointB.ChannelID,
		Bech32ValidatorAddrPrefix: params.Bech32PrefixValAddr,
		Validators:                selectedVals,
		Redemptionratio:           sdk.NewDec(1),
		IbcDenom: suite.calcuateIBCDenom(
			suite.transferPath.EndpointB.ChannelConfig.PortID,
			suite.transferPath.EndpointB.ChannelID,
			params.DefaultBondDenom),
		NativeDenom:     params.DefaultBondDenom,
		DerivativeDenom: "DLST",
	}

	return &sourceChain
}

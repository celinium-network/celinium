package keeper_test

import (
	"math/rand"

	params "github.com/celinium-network/celinium/app/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestAddSourceChain() {
	sourceChain := suite.mockSourceChainParams()

	ctlChainApp := getCeliniumApp(suite.controlChain)
	ctx := suite.controlChain.GetContext()
	ctlChainApp.EpochsKeeper.SetEpochInfo(ctx, *suite.delegationEpoch())

	channelSequence := ctlChainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := ctlChainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), sourceChain)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	createdICAs := getCreatedICAFromSourceChain(sourceChain)

	for _, ica := range createdICAs {
		suite.relayICACreatedPacket(channelSequence, ica)
		channelSequence++
	}
}

func getCreatedICAFromSourceChain(s *types.SourceChain) []string {
	return []string{s.WithdrawAddress, s.DelegateAddress}
}

func (suite *KeeperTestSuite) mockSourceChainParams() *types.SourceChain {
	sourceChainVals := len(suite.sourceChain.Vals.Validators)
	randVals := rand.Int()%sourceChainVals + 1 //nolint:gosec
	selectedVals := make([]types.Validator, 0)
	maxWeight := uint64(100000)

	selectedVals = append(selectedVals, types.Validator{
		Address: sdk.ValAddress(suite.sourceChain.Vals.Proposer.Address).String(),
		Weight:  rand.Uint64()%maxWeight + types.MinValidatorWeight, //nolint:gosec
	})

	for i := 0; i < randVals; i++ {
		vaddr := sdk.ValAddress(suite.sourceChain.Vals.Validators[i].Address).String()
		if vaddr == selectedVals[0].Address {
			continue
		}
		selectedVals = append(selectedVals, types.Validator{
			Address: vaddr,
			Weight:  rand.Uint64()%maxWeight + types.MinValidatorWeight, //nolint:gosec
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

func (suite *KeeperTestSuite) relayICACreatedPacket(packetSequence uint64, ctlAccount string) {
	controlChainApp := getCeliniumApp(suite.controlChain)
	sourceChainApp := getCeliniumApp(suite.sourceChain)

	portID, err := icatypes.NewControllerPortID(ctlAccount)
	suite.NoError(err)
	suite.icaPath.EndpointB.ChannelConfig.PortID = portID
	suite.icaPath.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(packetSequence)
	suite.icaPath.EndpointA.ChannelID = suite.icaPath.EndpointB.ChannelID

	channel, _ := controlChainApp.IBCKeeper.ChannelKeeper.GetChannel(
		suite.controlChain.GetContext(), portID, suite.icaPath.EndpointB.ChannelID)
	suite.icaPath.EndpointB.ChannelConfig.Version = channel.Version

	err = suite.icaPath.EndpointA.ChanOpenTry()
	suite.NoError(err)

	err = suite.icaPath.EndpointB.ChanOpenAck()
	suite.NoError(err)

	err = suite.icaPath.EndpointA.ChanOpenConfirm()
	suite.NoError(err)

	icaFromCtlChain, found := controlChainApp.ICAControllerKeeper.GetInterchainAccountAddress(
		suite.controlChain.GetContext(), suite.icaPath.EndpointB.ConnectionID, portID)

	suite.True(found)

	icaFromSrcChain, found := sourceChainApp.ICAHostKeeper.GetInterchainAccountAddress(
		suite.sourceChain.GetContext(), suite.icaPath.EndpointA.ConnectionID, portID)

	suite.True(found)
	suite.Equal(icaFromCtlChain, icaFromSrcChain)
}

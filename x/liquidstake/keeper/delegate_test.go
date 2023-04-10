package keeper_test

import (
	"time"

	epochtypes "github.com/celinium-netwok/celinium/x/epochs/types"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestCreateNewDelegationRecordAtEpochStart() {
	controlChainApp := getCeliniumApp(suite.controlChain)

	sourceChainParams := suite.mockSourceChainParams()
	channelSequence := controlChainApp.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(suite.controlChain.GetContext())

	err := controlChainApp.LiquidStakeKeeper.AddSouceChain(suite.controlChain.GetContext(), sourceChainParams)
	suite.NoError(err)
	suite.controlChain.NextBlock()

	createdICAs := getCreatedICAFromSourceChain(sourceChainParams)
	for _, ica := range createdICAs {
		suite.relayICACreatedPacket(channelSequence, ica)
		channelSequence++
	}

	// set delegation Epoch Info
	epochInfo := epochtypes.EpochInfo{
		Identifier:              types.DelegationEpochIdentifier,
		StartTime:               suite.coordinator.CurrentTime,
		Duration:                time.Hour,
		CurrentEpoch:            1,
		CurrentEpochStartTime:   suite.coordinator.CurrentTime,
		EpochCountingStarted:    true, // already start
		CurrentEpochStartHeight: suite.controlChain.GetContext().BlockHeight(),
	}

	controlChainApp.EpochsKeeper.SetEpochInfo(suite.controlChain.GetContext(), epochInfo)

	// start epoch and update off chain light.
	suite.controlChain.Coordinator.IncrementTimeBy(time.Hour)
	suite.transferPath.EndpointA.UpdateClient()

	// check new delegation record
	nextDelegationRecordID := controlChainApp.LiquidStakeKeeper.GetDelegationRecordID(suite.controlChain.GetContext())
	createdDelegationRecord, found := controlChainApp.LiquidStakeKeeper.GetDelegationRecord(suite.controlChain.GetContext(), nextDelegationRecordID-1)
	suite.True(found)
	suite.Equal(createdDelegationRecord, types.DelegationRecord{
		Id:             0,
		DelegationCoin: sdk.NewCoin(sourceChainParams.NativeDenom, sdk.ZeroInt()),
		Status:         types.DelegationPending,
		EpochNumber:    uint64(epochInfo.CurrentEpoch + 1),
		ChainID:        sourceChainParams.ChainID,
	})
}

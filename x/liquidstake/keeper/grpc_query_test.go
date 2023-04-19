package keeper_test

import (
	"context"
	"fmt"
	"strings"

	appparams "github.com/celinium-netwok/celinium/app/params"
	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestGRPCQuerySourceChain() {
	sourceChainParams := suite.generateSourceChainParams()
	suite.setSourceChainAndEpoch(sourceChainParams, suite.delegationEpoch())

	queryClient := suite.queryClient
	var req *types.QuerySourceChainRequest

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QuerySourceChainResponse)
		expErr    bool
	}{
		{
			"query exist source chain",
			func() {
				req = &types.QuerySourceChainRequest{
					ChainID: sourceChainParams.ChainID,
				}
			},
			func(suite *KeeperTestSuite, response *types.QuerySourceChainResponse) {
				suite.True(strings.Compare(response.SourceChain.ChainID, req.ChainID) == 0)
			},
			false,
		},
		{
			"query unknown source chain",
			func() {
				req = &types.QuerySourceChainRequest{
					ChainID: "noexist",
				}
			},
			func(suite *KeeperTestSuite, response *types.QuerySourceChainResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.SourceChain(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegation() {
	sourceChainParams := suite.generateSourceChainParams()
	suite.setSourceChainAndEpoch(sourceChainParams, suite.delegationEpoch())

	var req *types.QueryChainEpochDelegationRecordRequest

	ctx := suite.controlChain.GetContext()
	user := suite.controlChain.SenderAccount.GetAddress()
	controlChainApp := getCeliniumApp(suite.controlChain)
	delegateEpoch, _ := controlChainApp.EpochsKeeper.GetEpochInfo(ctx, appparams.DelegationEpochIdentifier)
	controlChainApp.LiquidStakeKeeper.Delegate(ctx, sourceChainParams.ChainID, suite.testCoin.Amount, user)
	suite.advanceEpochAndRelayIBC(suite.delegationEpoch())

	queryClient := suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QueryChainEpochDelegationRecordResponse)
		expErr    bool
	}{
		{
			"real",
			func() {
				req = &types.QueryChainEpochDelegationRecordRequest{
					ChainID: sourceChainParams.ChainID,
					Epoch:   uint64(delegateEpoch.CurrentEpoch),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochDelegationRecordResponse) {
				suite.True(strings.Compare(req.ChainID, response.Record.ChainID) == 0)
				suite.Equal(req.Epoch, response.Record.EpochNumber)
			},
			false,
		},
		{
			"query unknown source chain",
			func() {
				req = &types.QueryChainEpochDelegationRecordRequest{
					ChainID: "noexist",
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochDelegationRecordResponse) {},
			true,
		},
		{
			"query future epoch",
			func() {
				req = &types.QueryChainEpochDelegationRecordRequest{
					ChainID: sourceChainParams.ChainID,
					Epoch:   uint64(10000),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochDelegationRecordResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.ChainEpochDelegationRecord(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChainUnbondings() {
	srcChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, delegationEpochInfo)

	ctlChainApp := getCeliniumApp(suite.controlChain)
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()

	ctx := suite.controlChain.GetContext()
	ctlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID, suite.testCoin.Amount, ctlChainUserAccAddr)
	suite.advanceEpochAndRelayIBC(delegationEpochInfo)

	unbondingEpochInfo := suite.unbondEpoch()
	ctx = suite.controlChain.GetContext()
	ctlChainApp.EpochsKeeper.SetEpochInfo(ctx, *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	ctx = suite.controlChain.GetContext()
	ctlChainApp.LiquidStakeKeeper.Undelegate(ctx, srcChainParams.ChainID, suite.testCoin.Amount, ctlChainUserAccAddr)

	suite.advanceEpochAndRelayIBC(unbondingEpochInfo)

	var req *types.QueryChainEpochUnbondingRequest
	queryClient := suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QueryChainEpochUnbondingResponse)
		expErr    bool
	}{
		{
			"successful query",
			func() {
				req = &types.QueryChainEpochUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					Epoch:   uint64(2),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochUnbondingResponse) {
				suite.True(strings.Compare(req.ChainID, response.ChainUnbonding.ChainID) == 0)
			},
			false,
		},
		{
			"query unknown source chain",
			func() {
				req = &types.QueryChainEpochUnbondingRequest{
					ChainID: "noexist",
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochUnbondingResponse) {},
			true,
		},
		{
			"query future epoch",
			func() {
				req = &types.QueryChainEpochUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					Epoch:   uint64(10000),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryChainEpochUnbondingResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.ChainEpochUnbonding(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUserUndelegationRecord() {
	srcChainParams := suite.generateSourceChainParams()
	delegationEpochInfo := suite.delegationEpoch()
	suite.setSourceChainAndEpoch(srcChainParams, delegationEpochInfo)

	ctlChainApp := getCeliniumApp(suite.controlChain)
	ctlChainUserAccAddr := suite.controlChain.SenderAccount.GetAddress()
	unknownUser := suite.sourceChain.SenderAccount.GetAddress()

	ctx := suite.controlChain.GetContext()
	ctlChainApp.LiquidStakeKeeper.Delegate(ctx, srcChainParams.ChainID, suite.testCoin.Amount, ctlChainUserAccAddr)
	suite.advanceEpochAndRelayIBC(delegationEpochInfo)

	unbondingEpochInfo := suite.unbondEpoch()
	ctx = suite.controlChain.GetContext()
	ctlChainApp.EpochsKeeper.SetEpochInfo(ctx, *unbondingEpochInfo)
	suite.controlChain.Coordinator.IncrementTimeBy(unbondingEpochInfo.Duration)
	suite.transferPath.EndpointA.UpdateClient()

	ctx = suite.controlChain.GetContext()
	ctlChainApp.LiquidStakeKeeper.Undelegate(ctx, srcChainParams.ChainID, suite.testCoin.Amount, ctlChainUserAccAddr)

	suite.advanceEpochAndRelayIBC(unbondingEpochInfo)

	var req *types.QueryUserUndelegationRecordRequest
	queryClient := suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QueryUserUndelegationRecordResponse)
		expErr    bool
	}{
		{
			"successful query",
			func() {
				req = &types.QueryUserUndelegationRecordRequest{
					ChainID: srcChainParams.ChainID,
					User:    ctlChainUserAccAddr.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUndelegationRecordResponse) {
				suite.True(strings.Compare(req.ChainID, response.UndelegationRecords[0].ChainID) == 0)
				suite.True(strings.Contains(response.UndelegationRecords[0].ID, req.User))
			},
			false,
		},
		{
			"query user has't undelegation operation",
			func() {
				req = &types.QueryUserUndelegationRecordRequest{
					ChainID: srcChainParams.ChainID,
					User:    unknownUser.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUndelegationRecordResponse) {
				suite.True(len(response.UndelegationRecords) == 0)
			},
			false,
		},
		{
			"query unknown chain",
			func() {
				req = &types.QueryUserUndelegationRecordRequest{
					ChainID: "noexist",
					User:    ctlChainUserAccAddr.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUndelegationRecordResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.UserUndelegationRecord(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

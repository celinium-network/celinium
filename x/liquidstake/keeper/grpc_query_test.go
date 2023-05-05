package keeper_test

import (
	"context"
	"fmt"
	"strings"

	appparams "github.com/celinium-network/celinium/app/params"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

func (suite *KeeperTestSuite) TestGRPCQuerySourceChain() {
	sourceChainParams := suite.mockSourceChainParams()
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
	sourceChainParams := suite.mockSourceChainParams()
	suite.setSourceChainAndEpoch(sourceChainParams, suite.delegationEpoch())

	var req *types.QueryProxyDelegationRequest

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
		onSuccess func(suite *KeeperTestSuite, response *types.QueryProxyDelegationResponse)
		expErr    bool
	}{
		{
			"real",
			func() {
				req = &types.QueryProxyDelegationRequest{
					ChainID: sourceChainParams.ChainID,
					Epoch:   uint64(delegateEpoch.CurrentEpoch),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryProxyDelegationResponse) {
				suite.True(strings.Compare(req.ChainID, response.Record.ChainID) == 0)
				suite.Equal(req.Epoch, response.Record.EpochNumber)
			},
			false,
		},
		{
			"query unknown source chain",
			func() {
				req = &types.QueryProxyDelegationRequest{
					ChainID: "noexist",
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryProxyDelegationResponse) {},
			true,
		},
		{
			"query future epoch",
			func() {
				req = &types.QueryProxyDelegationRequest{
					ChainID: sourceChainParams.ChainID,
					Epoch:   uint64(10000),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryProxyDelegationResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.ProxyDelegation(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryChainUnbondings() {
	srcChainParams := suite.mockSourceChainParams()
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

	var req *types.QueryEpochProxyUnbondingRequest
	queryClient := suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QueryEpochProxyUnbondingResponse)
		expErr    bool
	}{
		{
			"successful query",
			func() {
				req = &types.QueryEpochProxyUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					Epoch:   uint64(2),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryEpochProxyUnbondingResponse) {
				suite.True(strings.Compare(req.ChainID, response.ChainUnbonding.ChainID) == 0)
			},
			false,
		},
		{
			"query unknown source chain",
			func() {
				req = &types.QueryEpochProxyUnbondingRequest{
					ChainID: "noexist",
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryEpochProxyUnbondingResponse) {},
			true,
		},
		{
			"query future epoch",
			func() {
				req = &types.QueryEpochProxyUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					Epoch:   uint64(10000),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryEpochProxyUnbondingResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.EpochProxyUnbonding(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryUserUnbonding() {
	srcChainParams := suite.mockSourceChainParams()
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

	var req *types.QueryUserUnbondingRequest
	queryClient := suite.queryClient

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(suite *KeeperTestSuite, response *types.QueryUserUnbondingResponse)
		expErr    bool
	}{
		{
			"successful query",
			func() {
				req = &types.QueryUserUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					User:    ctlChainUserAccAddr.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUnbondingResponse) {
				suite.True(strings.Compare(req.ChainID, response.UserUnbondings[0].ChainID) == 0)
				suite.True(strings.Contains(response.UserUnbondings[0].ID, req.User))
			},
			false,
		},
		{
			"query user has't undelegation operation",
			func() {
				req = &types.QueryUserUnbondingRequest{
					ChainID: srcChainParams.ChainID,
					User:    unknownUser.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUnbondingResponse) {
				suite.True(len(response.UserUnbondings) == 0)
			},
			false,
		},
		{
			"query unknown chain",
			func() {
				req = &types.QueryUserUnbondingRequest{
					ChainID: "noexist",
					User:    ctlChainUserAccAddr.String(),
				}
			},
			func(suite *KeeperTestSuite, response *types.QueryUserUnbondingResponse) {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := queryClient.UserUnbonding(context.Background(), req)
			if tc.expErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
				tc.onSuccess(suite, res)
			}
		})
	}
}

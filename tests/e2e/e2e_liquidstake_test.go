package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/celinium-network/celinium/app/params"
	liquistakecli "github.com/celinium-network/celinium/x/liquidstake/client/cli"
	"github.com/celinium-network/celinium/x/liquidstake/types"
)

type sourceChainParams struct {
	ChainID         string
	ConnectionID    string
	ChannelID       string
	ValPrefix       string
	CliVals         liquistakecli.CliValidators
	NativeDeonm     string
	DerivativeDenom string
	registor        string
}

func (s *IntegrationTestSuite) mockRegisterSourceChain() *sourceChainParams {
	regparams := sourceChainParams{
		ChainID:         s.srcChain.ID,
		ConnectionID:    "connection-0",
		ChannelID:       "channel-0",
		ValPrefix:       "celivaloper",
		NativeDeonm:     "CELI",
		DerivativeDenom: "vpCELI",
	}

	registor, err := s.ctlChain.validators[0].keyRecord.GetAddress()
	s.NoError(err)

	regparams.registor = registor.String()

	for _, v := range s.srcChain.validators {
		accAddr, _ := v.keyRecord.GetAddress()
		valAddr := sdk.ValAddress(accAddr)

		regparams.CliVals.Vals = append(regparams.CliVals.Vals, types.Validator{
			Address:          valAddr.String(),
			DelegationAmount: math.ZeroInt(),
			Weight:           1000000,
		})
	}

	return &regparams
}

func (s *IntegrationTestSuite) LiquidStakeAddSourceChain(regparams *sourceChainParams) (*types.SourceChain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	selectValsBz, err := json.Marshal(regparams.CliVals)
	if err != nil {
		return nil, err
	}

	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)
	liuidstakeCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"register-source-chain",
		regparams.ChainID,
		regparams.ConnectionID,
		regparams.ChannelID,
		regparams.ValPrefix,
		string(selectValsBz),
		regparams.NativeDeonm,
		regparams.DerivativeDenom,
		fmt.Sprintf("--from=%s", regparams.registor),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, gas*10),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.ctlChain.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))

	resp, err := queryLiquidstakeSourceChain(s.ctlChain.encfg.Codec, chainBAPIEndpoint, s.srcChain.ID)
	if err != nil {
		return nil, err
	}

	// TODO check registe result ?
	return &resp.SourceChain, nil
}

func (s *IntegrationTestSuite) LiquistakeDelegate(sourceChain *types.SourceChain, amount math.Int) {
	address, _ := s.srcChain.validators[0].keyRecord.GetAddress()
	srcUser := address.String()
	address, _ = s.ctlChain.validators[0].keyRecord.GetAddress()
	ctlUser := address.String()
	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))

	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)
	s.sendIBC(s.srcChain, 0, srcUser, ctlUser, amount.String()+s.srcChain.Denom, fee.String(), "")

	time.Sleep(time.Second * 50)

	curDelegationEpochResp, err := queryCurEpoch(s.ctlChain.encfg.Codec, chainBAPIEndpoint, appparams.DelegationEpochIdentifier)
	curEpoch := uint64(curDelegationEpochResp.CurrentEpoch)
	s.NoError(err)

	liuidstakeDelegateCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"delegate",
		sourceChain.ChainID,
		amount.String(),
		fmt.Sprintf("--from=%s", ctlUser),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, gas*10),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.ctlChain.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcBalBefore, err := getSpecificBalance(s.ctlChain.encfg.Codec, chainBAPIEndpoint, ctlUser, sourceChain.IbcDenom)
	s.NoError(err)

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeDelegateCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	ibcBalAfter, err := getSpecificBalance(s.ctlChain.encfg.Codec, chainBAPIEndpoint, ctlUser, sourceChain.IbcDenom)
	s.NoError(err)

	s.True(ibcBalBefore.Sub(ibcBalAfter).Amount.Equal(amount))
	resp, err := queryLiquidstakeDelegationRecord(s.ctlChain.encfg.Codec, chainBAPIEndpoint, s.srcChain.ID, curEpoch)
	s.NoError(err)

	targetDelegationRecord := types.DelegationRecord{
		DelegationCoin: sdk.Coin{
			Denom:  sourceChain.IbcDenom,
			Amount: amount,
		},
		Status:            types.DelegationPending,
		EpochNumber:       curEpoch,
		ChainID:           sourceChain.ChainID,
		TransferredAmount: math.ZeroInt(),
	}

	s.True(compareDelegationRecord(&resp.Record, &targetDelegationRecord))
}

func (s *IntegrationTestSuite) CheckChainDelegate(sourceChain *types.SourceChain) {
	// wait for next delegate epoch
	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))

	curDelegationEpochResp, err := queryCurEpoch(s.ctlChain.encfg.Codec, ctlAPIEndpoint, appparams.DelegationEpochIdentifier)
	s.NoError(err)

	for {
		resp, err := queryCurEpoch(s.ctlChain.encfg.Codec, ctlAPIEndpoint, appparams.DelegationEpochIdentifier)
		s.NoError(err)
		if curDelegationEpochResp.CurrentEpoch < resp.CurrentEpoch {
			break
		}
		time.Sleep(time.Second * 30)
	}

	time.Sleep(time.Minute)

	delegateICA, err := queryInterChainAccount(s.ctlChain.encfg.Codec, ctlAPIEndpoint,
		sourceChain.DelegateAddress, sourceChain.ConnectionID)
	s.NoError(err)

	totalDelegateAmt := math.ZeroInt()
	for _, v := range sourceChain.Validators {
		srcDelegation, err := queryDelegation(s.srcChain.encfg.Codec, srcAPIEndpoint, v.Address, delegateICA)
		s.NoError(err)
		totalDelegateAmt = totalDelegateAmt.Add(srcDelegation.DelegationResponse.Balance.Amount)
	}

	res, err := queryLiquidstakeSourceChain(s.srcChain.encfg.Codec, ctlAPIEndpoint, sourceChain.ChainID)

	s.NoError(err)
	s.True(res.SourceChain.StakedAmount.Equal(totalDelegateAmt))
}

func (s *IntegrationTestSuite) CheckChainReinvest(sourceChain *types.SourceChain) {
	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))

	epochInfo, err := getSpecicalEpochInfo(s.ctlChain.encfg.Codec, ctlAPIEndpoint, appparams.ReinvestEpochIdentifier)
	s.NoError(err)

	epochRemindingDuration := epochInfo.Duration - time.Since(epochInfo.CurrentEpochStartTime)
	time.Sleep(epochRemindingDuration - time.Second)

	delegateICA, err := queryInterChainAccount(s.ctlChain.encfg.Codec, ctlAPIEndpoint,
		sourceChain.DelegateAddress, sourceChain.ConnectionID)
	s.NoError(err)

	delegateReward := math.ZeroInt()
	for _, v := range sourceChain.Validators {
		rewardAmt, err := queryDelegationReward(s.srcChain.encfg.Codec, srcAPIEndpoint, delegateICA, v.Address)
		s.NoError(err)
		delegateReward = delegateReward.Add(rewardAmt)
	}
	fmt.Println(delegateReward)
}

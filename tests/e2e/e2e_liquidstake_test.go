package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
			Address:     valAddr.String(),
			TokenAmount: math.ZeroInt(),
			Weight:      1000000,
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

	s.Logf("Liquistake regisor source chain, chainID %s", regparams.ChainID)
	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))

	resp, err := queryLiquidstakeSourceChain(s.ctlChain.encfg.Codec, chainBAPIEndpoint, s.srcChain.ID)
	if err != nil {
		return nil, err
	}
	s.Logf("Liquistake regisor source chain successful, chainID %s", regparams.ChainID)
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

	s.Logf("Liquistake begin delegate, chainID %s, amount %s ,user: %s , epoch %d",
		sourceChain.ChainID, amount.String(), ctlUser, curEpoch)

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeDelegateCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	ibcBalAfter, err := getSpecificBalance(s.ctlChain.encfg.Codec, chainBAPIEndpoint, ctlUser, sourceChain.IbcDenom)
	s.NoError(err)

	s.True(ibcBalBefore.Sub(ibcBalAfter).Amount.Equal(amount))
	resp, err := queryLiquidstakeDelegation(s.ctlChain.encfg.Codec, chainBAPIEndpoint, s.srcChain.ID, curEpoch)
	s.NoError(err)

	targetProxyDelegation := types.ProxyDelegation{
		Coin: sdk.Coin{
			Denom:  sourceChain.IbcDenom,
			Amount: amount,
		},
		Status:            types.ProxyDelegationPending,
		EpochNumber:       curEpoch,
		ChainID:           sourceChain.ChainID,
		TransferredAmount: math.ZeroInt(),
	}

	s.True(compareProxyDelegation(&resp.Record, &targetProxyDelegation))
	s.Logf("Liquistake begin delegate successful")
}

func (s *IntegrationTestSuite) CheckChainDelegate(sourceChain *types.SourceChain) {
	s.Logf("Begin check chain Delegation")

	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))

	s.waitForNextEpoch(ctlAPIEndpoint, appparams.DelegationEpochIdentifier, time.Second*15)

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

	s.Logf("Liquidstake Chain Delegation Successful. SourceChain %v", res.SourceChain)
}

func (s *IntegrationTestSuite) LiquidstakeReinvest(sourceChain *types.SourceChain) math.Int {
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
	s.Logf("Begin reinvest, reward %s", delegateReward.String())

	s.waitForNextEpoch(ctlAPIEndpoint, appparams.ReinvestEpochIdentifier, time.Second*5)
	time.Sleep(time.Second * 10)

	// The current state should be as follows:
	// 1) The withdrawal account of delegation has been set to the interchain account corresponding to
	//	  the withdrawal address in the source chain.
	// 2) The withdrawal account withdraws the reward and transfers it to the delegate address.
	// 3) The ProxyDelegation belonging to the current epoch has recorded the transferred reward funds.

	/* begin check */
	resp, err := queryCurEpoch(s.ctlChain.encfg.Codec, ctlAPIEndpoint, appparams.DelegationEpochIdentifier)
	s.NoError(err)

	// check trannferred amount
	rcResp, err := queryLiquidstakeDelegation(s.ctlChain.encfg.Codec, ctlAPIEndpoint, sourceChain.ChainID, uint64(resp.CurrentEpoch))
	s.NoError(err)
	s.True(rcResp.Record.TransferredAmount.GT(delegateReward))

	// check withdraw address has been correctly setted.
	withdrawAddressResp, err := queryDelegatorWithdrawalAddress(s.srcChain.encfg.Codec, srcAPIEndpoint, delegateICA)
	s.NoError(err)
	witdrawICA, err := queryInterChainAccount(s.ctlChain.encfg.Codec, ctlAPIEndpoint,
		sourceChain.WithdrawAddress, sourceChain.ConnectionID)
	s.NoError(err)
	s.Equal(withdrawAddressResp.WithdrawAddress, witdrawICA)

	// check all reward has transferred to delegatorICA and correctly record.
	balance, err := getSpecificBalance(s.srcChain.encfg.Codec, srcAPIEndpoint, delegateICA, sourceChain.NativeDenom)
	s.NoError(err)
	s.True(rcResp.Record.TransferredAmount.Equal(balance.Amount))

	s.Logf("Reinvest successfully")
	return balance.Amount
}

func (s *IntegrationTestSuite) CheckChainReinvest(srcChain *types.SourceChain, delegateAmount, rewadAmount math.Int) {
	// wait for next delegation epoch
	// 1) check redeem rate
	// 2) check delegation in source chain
	// 3) check source chain stakedamount

	s.Logf("Begin check reinvest")

	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))

	s.waitForNextEpoch(ctlAPIEndpoint, appparams.DelegationEpochIdentifier, time.Second*10)
	time.Sleep(time.Second * 10)

	res, err := queryLiquidstakeSourceChain(s.srcChain.encfg.Codec, ctlAPIEndpoint, srcChain.ChainID)
	s.NoError(err)

	s.True(res.SourceChain.StakedAmount.Equal(delegateAmount.Add(rewadAmount)))

	delegateICA, err := queryInterChainAccount(s.ctlChain.encfg.Codec, ctlAPIEndpoint,
		srcChain.DelegateAddress, srcChain.ConnectionID)
	s.NoError(err)

	totalDelegateAmt := math.ZeroInt()
	for _, v := range srcChain.Validators {
		srcDelegation, err := queryDelegation(s.srcChain.encfg.Codec, srcAPIEndpoint, v.Address, delegateICA)
		s.NoError(err)
		totalDelegateAmt = totalDelegateAmt.Add(srcDelegation.DelegationResponse.Balance.Amount)
	}
	s.True(res.SourceChain.StakedAmount.Equal(totalDelegateAmt))

	rate := sdk.NewDecFromInt(res.SourceChain.StakedAmount).QuoInt(delegateAmount)
	s.True(res.SourceChain.Redemptionratio.Equal(rate))

	s.Logf("Check reinvest successful")
}

func (s *IntegrationTestSuite) LiquistakeUndelegate(srcChain *types.SourceChain, undelegateAmount math.Int, rewardAmount math.Int) uint64 {
	s.Logf("Begin LiquistakeUndelegate ...")

	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	// srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))
	address, _ := s.ctlChain.validators[0].keyRecord.GetAddress()
	ctlUser := address.String()
	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)

	liuidstakeUndelegateCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"undelegate",
		srcChain.ChainID,
		undelegateAmount.String(),
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

	ibcBalBefore, err := getSpecificBalance(s.ctlChain.encfg.Codec, ctlAPIEndpoint, ctlUser, srcChain.DerivativeDenom)
	s.NoError(err)

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeUndelegateCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	ibcBalAfter, err := getSpecificBalance(s.ctlChain.encfg.Codec, ctlAPIEndpoint, ctlUser, srcChain.DerivativeDenom)
	s.NoError(err)

	// derivative denom should be reduce from the caller
	s.True(ibcBalBefore.Amount.Sub(ibcBalAfter.Amount).Equal(undelegateAmount))

	epochRes, err := queryCurEpoch(s.ctlChain.encfg.Codec, ctlAPIEndpoint, appparams.UndelegationEpochIdentifier)
	s.NoError(err)

	chainUnbondingResp, err := queryLiquidstakeProxyUnbonding(s.ctlChain.encfg.Codec, ctlAPIEndpoint, srcChain.ChainID, uint64(epochRes.CurrentEpoch))
	s.NoError(err)
	fmt.Println(chainUnbondingResp)

	s.True(chainUnbondingResp.ChainUnbonding.BurnedDerivativeAmount.Equal(undelegateAmount))
	s.True(chainUnbondingResp.ChainUnbonding.RedeemNativeToken.Amount.Equal(undelegateAmount.Add(rewardAmount)))

	// check user undelegate reocrd
	userUnbondingResp, err := queryLiquidstakeUserUnbonding(s.ctlChain.encfg.Codec, ctlAPIEndpoint, srcChain.ChainID, ctlUser)
	s.NoError(err)
	for _, rc := range userUnbondingResp.UserUnbondings {
		if rc.Epoch == uint64(epochRes.CurrentEpoch) {
			s.True(rc.RedeemCoin.Amount.Equal(undelegateAmount.Add(rewardAmount)))
			s.Equal(rc.CliamStatus, types.UserUnbondingPending)
		}
	}

	s.waitForNextEpoch(ctlAPIEndpoint, appparams.UndelegationEpochIdentifier, time.Second*10)
	chainUnbondingResp, _ = queryLiquidstakeProxyUnbonding(s.ctlChain.encfg.Codec, ctlAPIEndpoint, srcChain.ChainID, uint64(epochRes.CurrentEpoch))
	completeTimeStamp := chainUnbondingResp.ChainUnbonding.UnbondTime
	completeTime := time.Unix(0, int64(completeTimeStamp))

	s.Logf("Undelegate complete time %s", completeTime.Format(time.RFC3339))

	time.Sleep(time.Until(completeTime) + time.Minute*2)

	s.waitForNextEpoch(ctlAPIEndpoint, appparams.UndelegationEpochIdentifier, time.Second*20)
	time.Sleep(time.Minute)

	userUnbondingResp, err = queryLiquidstakeUserUnbonding(s.ctlChain.encfg.Codec, ctlAPIEndpoint, srcChain.ChainID, ctlUser)
	s.NoError(err)
	for _, rc := range userUnbondingResp.UserUnbondings {
		if rc.Epoch == uint64(epochRes.CurrentEpoch) {
			s.True(rc.RedeemCoin.Amount.Equal(undelegateAmount.Add(rewardAmount)))
			s.Equal(rc.CliamStatus, types.UserUnbondingClaimable)
		}
	}

	return uint64(epochRes.CurrentEpoch)
}

func (s *IntegrationTestSuite) LiquidstakeClaim(claimableCoin sdk.Coin, chainID string, epoch uint64) {
	ctlAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.ctlChain.ID][0].GetHostPort("1317/tcp"))
	// srcAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.srcChain.ID][0].GetHostPort("1317/tcp"))
	address, _ := s.ctlChain.validators[0].keyRecord.GetAddress()
	ctlUser := address.String()
	fee := sdk.NewCoin(s.srcChain.Denom, standardFeeAmount)

	liuidstakeClaimCmd := []string{
		s.ctlChain.ChainNodeBinary,
		txCommand,
		"liquidstake",
		"claim",
		chainID,
		strconv.Itoa(int(epoch)),
		fmt.Sprintf("--from=%s", ctlUser),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, gas*10),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.ctlChain.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.Logf("Begin Liquidstake Claim, amount %s", claimableCoin.Amount.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcBalBefore, err := getSpecificBalance(s.ctlChain.encfg.Codec, ctlAPIEndpoint, ctlUser, claimableCoin.Denom)
	s.NoError(err)

	s.executeCeliniumTxCommand(ctx, s.ctlChain, liuidstakeClaimCmd, 0, s.defaultExecValidation(s.ctlChain, 0))
	ibcBalAfter, err := getSpecificBalance(s.ctlChain.encfg.Codec, ctlAPIEndpoint, ctlUser, claimableCoin.Denom)
	s.NoError(err)

	s.True(ibcBalAfter.Amount.Sub(ibcBalBefore.Amount).Equal(claimableCoin.Amount))

	s.Logf("Liquidstake claim successful %s", claimableCoin.Amount.String())
}

func (s *IntegrationTestSuite) waitForNextEpoch(endpoint, identifier string, interval time.Duration) {
	curResp, err := queryCurEpoch(s.ctlChain.encfg.Codec, endpoint, identifier)
	s.NoError(err)

	for {
		s.Logf("Waitting next %s epoch", identifier)
		resp, err := queryCurEpoch(s.ctlChain.encfg.Codec, endpoint, identifier)
		s.NoError(err)
		if curResp.CurrentEpoch < resp.CurrentEpoch {
			break
		}
		time.Sleep(time.Second * 30)
	}
	s.Logf("reaching the next %s epoch", identifier)
}

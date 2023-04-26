package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/types"

	epochtypes "github.com/celinium-network/celinium/x/epochs/types"
	liquidstaketypes "github.com/celinium-network/celinium/x/liquidstake/types"
)

func queryChainTx(endpoint, txHash string) error {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}

	return nil
}

// if coin is zero, return empty coin.
func getSpecificBalance(cdc codec.Codec, endpoint, addr, denom string) (amt sdk.Coin, err error) {
	balances, err := queryAllBalances(cdc, endpoint, addr)
	amt.Amount = math.ZeroInt()
	if err != nil {
		return amt, err
	}
	for _, c := range balances {
		if strings.Contains(c.Denom, denom) {
			amt = c
			break
		}
	}
	return amt, nil
}

func queryAllBalances(cdc codec.Codec, endpoint, addr string) (sdk.Coins, error) {
	body, err := httpGet(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := cdc.UnmarshalJSON(body, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func queryDelegation(cdc codec.Codec, endpoint string, validatorAddr string, delegatorAddr string) (stakingtypes.QueryDelegationResponse, error) {
	var res stakingtypes.QueryDelegationResponse

	body, err := httpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", endpoint, validatorAddr, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryDelegatorWithdrawalAddress(cdc codec.Codec, endpoint string, delegatorAddr string) (disttypes.QueryDelegatorWithdrawAddressResponse, error) {
	var res disttypes.QueryDelegatorWithdrawAddressResponse

	body, err := httpGet(fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", endpoint, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryAccount(cdc codec.Codec, endpoint, address string) (acc authtypes.AccountI, err error) { //nolint:unused // this is called during e2e tests
	var res authtypes.QueryAccountResponse
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", endpoint, address))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := cdc.UnmarshalJSON(bz, &res); err != nil {
		return nil, err
	}
	return acc, cdc.UnpackAny(res.Account, &acc)
}

func queryDelayedVestingAccount(cdc codec.Codec, endpoint, address string) (authvesting.DelayedVestingAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(cdc, endpoint, address)
	if err != nil {
		return authvesting.DelayedVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.DelayedVestingAccount)
	if !ok {
		return authvesting.DelayedVestingAccount{},
			fmt.Errorf("cannot cast %v to DelayedVestingAccount", baseAcc)
	}
	return *acc, nil
}

func queryContinuousVestingAccount(cdc codec.Codec, endpoint, address string) (authvesting.ContinuousVestingAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(cdc, endpoint, address)
	if err != nil {
		return authvesting.ContinuousVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.ContinuousVestingAccount)
	if !ok {
		return authvesting.ContinuousVestingAccount{},
			fmt.Errorf("cannot cast %v to ContinuousVestingAccount", baseAcc)
	}
	return *acc, nil
}

func queryPermanentLockedAccount(cdc codec.Codec, endpoint, address string) (authvesting.PermanentLockedAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(cdc, endpoint, address)
	if err != nil {
		return authvesting.PermanentLockedAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.PermanentLockedAccount)
	if !ok {
		return authvesting.PermanentLockedAccount{},
			fmt.Errorf("cannot cast %v to PermanentLockedAccount", baseAcc)
	}
	return *acc, nil
}

func queryPeriodicVestingAccount(cdc codec.Codec, endpoint, address string) (authvesting.PeriodicVestingAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(cdc, endpoint, address)
	if err != nil {
		return authvesting.PeriodicVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.PeriodicVestingAccount)
	if !ok {
		return authvesting.PeriodicVestingAccount{},
			fmt.Errorf("cannot cast %v to PeriodicVestingAccount", baseAcc)
	}
	return *acc, nil
}

func queryValidator(cdc codec.Codec, endpoint, address string) (stakingtypes.Validator, error) { //nolint:unused // this is called during e2e tests
	var res stakingtypes.QueryValidatorResponse

	body, err := httpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", endpoint, address))
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return stakingtypes.Validator{}, err
	}
	return res.Validator, nil
}

func queryValidators(cdc codec.Codec, endpoint string) (stakingtypes.Validators, error) { //nolint:unused // this is called during e2e tests
	var res stakingtypes.QueryValidatorsResponse
	body, err := httpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return nil, err
	}
	return res.Validators, nil
}

func queryEvidence(cdc codec.Codec, endpoint, hash string) (evidencetypes.QueryEvidenceResponse, error) { //nolint:unused // this is called during e2e tests
	var res evidencetypes.QueryEvidenceResponse
	body, err := httpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryAllEvidence(cdc codec.Codec, endpoint string) (evidencetypes.QueryAllEvidenceResponse, error) { //nolint:unused // this is called during e2e tests
	var res evidencetypes.QueryAllEvidenceResponse
	body, err := httpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence", endpoint))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryLiquidstakeSourceChain(cdc codec.Codec, endpoint string, chainID string) (liquidstaketypes.QuerySourceChainResponse, error) {
	var res liquidstaketypes.QuerySourceChainResponse
	body, err := httpGet(fmt.Sprintf("%s/celinium/liquidstake/v1/source_chain?ChainID=%s", endpoint, chainID))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryLiquidstakeDelegationRecord(cdc codec.Codec, endpoint string, chainID string, epoch uint64) (
	liquidstaketypes.QueryChainEpochDelegationRecordResponse, error,
) {
	var res liquidstaketypes.QueryChainEpochDelegationRecordResponse
	body, err := httpGet(fmt.Sprintf("%s/celinium/liquidstake/v1/chain_epoch_delegation?chainID=%s&epoch=%d", endpoint, chainID, epoch))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryLiquidstakeChainUnbonding(cdc codec.Codec, endpoint string, chainID string, epoch uint64) (
	liquidstaketypes.QueryChainEpochUnbondingResponse, error,
) {
	var res liquidstaketypes.QueryChainEpochUnbondingResponse
	body, err := httpGet(fmt.Sprintf("%s/celinium/liquidstake/v1/chain_epoch_unbonding?chainID=%s&epoch=%d", endpoint, chainID, epoch))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryLiquidstakeUserUnbonding(cdc codec.Codec, endpoint, chainID, userAddr string) (
	liquidstaketypes.QueryUserUndelegationRecordResponse, error,
) {
	var res liquidstaketypes.QueryUserUndelegationRecordResponse
	body, err := httpGet(fmt.Sprintf("%s/celinium/liquidstake/v1/user_undelegation_record?chainID=%s&user=%s", endpoint, chainID, userAddr))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryCurEpoch(cdc codec.Codec, endpoint string, identifier string) (epochtypes.QueryCurrentEpochResponse, error) {
	var res epochtypes.QueryCurrentEpochResponse

	body, err := httpGet(fmt.Sprintf("%s/celinium/epochs/v1/current_epoch?identifier=%s", endpoint, identifier))
	if err != nil {
		return res, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func getSpecicalEpochInfo(cdc codec.Codec, endpoint string, identifier string) (*epochtypes.EpochInfo, error) {
	var res epochtypes.QueryEpochsInfoResponse
	var epochInfo epochtypes.EpochInfo
	body, err := httpGet(fmt.Sprintf("%s/celinium/epochs/v1/epochs", endpoint))
	if err != nil {
		return nil, err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return nil, err
	}

	for _, e := range res.Epochs {
		if strings.Compare(e.Identifier, identifier) == 0 {
			epochInfo = e
		}
	}

	return &epochInfo, nil
}

func queryInterChainAccount(cdc codec.Codec, endpoint, owner, connectionID string) (string, error) {
	var res icacontrollertypes.QueryInterchainAccountResponse

	body, err := httpGet(fmt.Sprintf("%s/ibc/apps/interchain_accounts/controller/v1/owners/%s/connections/%s", endpoint, owner, connectionID))
	if err != nil {
		return "", err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return "", err
	}
	return res.Address, nil
}

func queryDelegationReward(cdc codec.Codec, endpoint, delegator, validator string) (math.Int, error) {
	var res disttypes.QueryDelegationRewardsResponse

	body, err := httpGet(fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", endpoint, delegator, validator))
	if err != nil {
		return math.ZeroInt(), err
	}

	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return math.ZeroInt(), err
	}
	return res.Rewards[0].Amount.TruncateInt(), nil
}

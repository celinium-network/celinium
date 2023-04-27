package e2e

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	rawTxFile = "tx_raw.json"
)

// buildRawTx build a dummy tx using the TxBuilder and
// return the JSON and encoded tx's
func buildRawTx(c *chain) ([]byte, string, error) {
	txConfig := c.encfg.TxConfig
	builder := txConfig.NewTxBuilder()

	builder.SetGasLimit(gas)
	builder.SetFeeAmount(sdk.Coins{sdk.NewCoin(c.Denom, standardFeeAmount)})
	builder.SetMemo("foomemo")

	tx, err := txConfig.TxJSONEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	txBytes, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	return tx, base64.StdEncoding.EncodeToString(txBytes), err
}

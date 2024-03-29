#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml

echo $CELI_SRC_E2E_RLY_MNEMONIC > /root/.hermes/CELI_SRC_E2E_RLY_MNEMONIC.txt
echo $CELI_CTL_E2E_RLY_MNEMONIC > /root/.hermes/CELI_CTL_E2E_RLY_MNEMONIC.txt

# setup Hermes relayer configuration
tee /root/.hermes/config.toml <<EOF
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3031

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001

[[chains]]
id = '$CELI_CTL_E2E_CHAIN_ID'
rpc_addr = 'http://$CELI_CTL_E2E_VAL_HOST:26657'
grpc_addr = 'http://$CELI_CTL_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$CELI_CTL_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'celi'
key_name = 'rly01-celi-ctl'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 0.00001, denom = 'CELI' }
gas_multiplier = 1.2
clock_drift = '1m' # to accomdate docker containers
trusting_period = '1hours'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '$CELI_SRC_E2E_CHAIN_ID'
rpc_addr = 'http://$CELI_SRC_E2E_VAL_HOST:26657'
grpc_addr = 'http://$CELI_SRC_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$CELI_SRC_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'celi'
key_name = 'rly01-celi-src'
store_prefix = 'ibc'
max_gas =  6000000
gas_price = { price = 0.00001, denom = 'CELI' }
gas_multiplier = 1.2
clock_drift = '1m' # to accomdate docker containers
trusting_period = '1hours'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

# import keys
hermes keys add  --key-name rly01-celi-src  --chain $CELI_SRC_E2E_CHAIN_ID --mnemonic-file /root/.hermes/CELI_SRC_E2E_RLY_MNEMONIC.txt
sleep 5
hermes keys add  --key-name rly01-celi-ctl  --chain $CELI_CTL_E2E_CHAIN_ID --mnemonic-file /root/.hermes/CELI_CTL_E2E_RLY_MNEMONIC.txt
sleep 5
# start Hermes relayer
hermes start

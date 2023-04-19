#!/bin/bash

set -o errexit -o nounset


_get_node_address() {
  $CHAIN_NODE keys list --output json | jq .[] | jq .address
}

_query_delegation() {
  $CHAIN_NODE query staking delegation $2 $3 --output json | jq .balance | jq .amount | sed 's/\"//g'
}

_get_wallet_balance() {
  $CHAIN_NODE query bank balances $WALLET_ADDRESS
}

_transfer() {
  $CHAIN_NODE tx bank send \
    $2 $3 $4 \
    --chain-id=mocha \
    --gas="auto" \
    --gas-adjustment=1.5 \
    --fees="5000$DENOM" \
    --from=$VALIDATOR_WALLET_NAME \
    --keyring-backend=test
}

if [ "$1" = 'wallet:balance' ]; then
  _get_wallet_balance
elif [ "$1" = 'wallet:address' ]; then
  _get_node_address  
elif [ "$1" = 'validator:query_delegation' ]; then
   _query_delegation "$@"
elif [ "$1" = 'wallet:transfer' ]; then
   _transfer "$@"
else
  $CHAIN_NODE "$@"
fi    
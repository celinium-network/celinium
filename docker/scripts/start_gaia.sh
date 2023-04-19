#!/bin/bash

set -o errexit -o nounset

WALLET_KEY_NAME=$VALIDATOR_NAME 
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000$DENOM"
STAKING_AMOUNT="100000000000000$DENOM"

gaiad tendermint unsafe-reset-all 
gaiad init $VALIDATOR_NAME --chain-id $CHAIN_ID 

sed -i "s/stake/$DENOM/g" ~/.gaia/config/genesis.json

gaiad keys add $WALLET_KEY_NAME --keyring-backend test 
gaiad add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test 

gaiad keys add gaia-0 --keyring-backend test 
gaiad add-genesis-account gaia-0 $TOKEN_AMOUNT --keyring-backend test 

gaiad gentx $WALLET_KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test 

gaiad collect-gentxs 

gaiad start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090 

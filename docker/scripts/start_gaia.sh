#!/bin/bash

set -o errexit -o nounset

VALIDATOR_NAME=TGAIA
CHAIN_ID=gaia
KEY_NAME=$VALIDATOR_NAME 
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000atom"
STAKING_AMOUNT="100000000000000atom"

gaiad tendermint unsafe-reset-all 
gaiad init $VALIDATOR_NAME --chain-id $CHAIN_ID 

sed -i 's/stake/atom/g' ~/.gaia/config/genesis.json

gaiad keys add $KEY_NAME --keyring-backend test 
gaiad add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test 

gaiad keys add gaia-0 --keyring-backend test 
gaiad add-genesis-account gaia-0 $TOKEN_AMOUNT --keyring-backend test 

gaiad gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test 

gaiad collect-gentxs 

gaiad start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090 

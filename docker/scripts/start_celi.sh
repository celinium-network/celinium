#!/bin/sh

VALIDATOR_NAME=validator1
CHAIN_ID=celinium
KEY_NAME=validator1   
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000CELI"
STAKING_AMOUNT="1000000000CELI"
NODEIP="--node http://127.0.0.1:26657"

celiniumd tendermint unsafe-reset-all
celiniumd init $VALIDATOR_NAME --chain-id $CHAIN_ID

celiniumd keys add $KEY_NAME --keyring-backend test
celiniumd add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test

celiniumd keys add celi-0 --keyring-backend test
celiniumd add-genesis-account celi-0 $TOKEN_AMOUNT --keyring-backend test

celiniumd gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

celiniumd collect-gentxs

celiniumd start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090

#!/bin/sh

WALLET_KEY_NAME=$VALIDATOR_NAME   
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000CELI"
STAKING_AMOUNT="1000000000CELI"
NODEIP="--node http://127.0.0.1:26657"

celiniumd tendermint unsafe-reset-all
celiniumd init $VALIDATOR_NAME --chain-id $CHAIN_ID

celiniumd keys add $WALLET_KEY_NAME --keyring-backend test
celiniumd add-genesis-account $WALLET_KEY_NAME $TOKEN_AMOUNT --keyring-backend test

celiniumd keys add celi-0 --keyring-backend test
celiniumd add-genesis-account celi-0 $TOKEN_AMOUNT --keyring-backend test

celiniumd gentx $WALLET_KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

celiniumd collect-gentxs

celiniumd start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address 0.0.0.0:9090

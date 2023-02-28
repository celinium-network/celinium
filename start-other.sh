#!/bin/sh

VALIDATOR_NAME=validator1
CHAIN_ID=celiniumd-other
KEY_NAME=$VALIDATOR_NAME 
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000CELI"
STAKING_AMOUNT="1000000000CELI"
OHTER_CHAIN_HOME=/home/manjaro/.celinium-other

./celiniumd tendermint unsafe-reset-all --home $OHTER_CHAIN_HOME
./celiniumd init $VALIDATOR_NAME --chain-id $CHAIN_ID --home $OHTER_CHAIN_HOME

# ./celiniumd keys add $KEY_NAME --keyring-backend test --home $OHTER_CHAIN_HOME
./celiniumd add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test --home $OHTER_CHAIN_HOME
./celiniumd gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test --home $OHTER_CHAIN_HOME
./celiniumd collect-gentxs --home $OHTER_CHAIN_HOME

./celiniumd start --rpc.laddr tcp://127.0.0.1:46659 --grpc.address 0.0.0.0:10091 --home $OHTER_CHAIN_HOME

# ./celiniumd tx bank send cosmos19mtmkgrlv9nd9u723hcrerfhfy4ph08q940pg2 \
#  cosmos1a4cj608rwutywqnxcd95cf8v923j9xhtuc48crekawe9f9tndy3qc6r2zw 1010010010CELI --fees 5000CELI --keyring-backend test --chain-id celiniumd-other --home /home/manjaro/.celinium-other

# key-name: ica-host-1 account cosmos1g39h56ugk78crm38tr0v0w49s2eflfuhxltsyx
# ica host account cosmos1a4cj608rwutywqnxcd95cf8v923j9xhtuc48crekawe9f9tndy3qc6r2zw
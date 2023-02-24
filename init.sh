#!/bin/sh

VALIDATOR_NAME=validator1
CHAIN_ID=celiniumd-5
KEY_NAME=ballman
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000demo"
STAKING_AMOUNT="1000000000demo"
NODEIP="--node http://127.0.0.1:26657"

# ./celiniumd tendermint unsafe-reset-all
# ./celiniumd init $VALIDATOR_NAME --chain-id $CHAIN_ID

# #  ./celiniumd keys add $KEY_NAME --keyring-backend test
# ./celiniumd add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test
# ./celiniumd gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test
# ./celiniumd collect-gentxs

./celiniumd start   --rpc.laddr tcp://127.0.0.1:46658 --grpc.address 0.0.0.0:10090 

# ./celiniumd tx bank send cosmos1slsftuhd9z8qql79v49e5yk8ehxnsjmagq3asw \
#  cosmos1r2mmcgghx4hzck4mezc7nlrylw99tuu7zx4x53 1000000000demo --fees 5000demo --keyring-backend test --chain-id celiniumd-4

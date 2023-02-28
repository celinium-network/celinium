#!/bin/sh

VALIDATOR_NAME=validator1
CHAIN_ID=celiniumd-5
KEY_NAME=validator1   
CHAINFLAG="--chain-id ${CHAIN_ID}"
TOKEN_AMOUNT="10000000000000000000000000CELI"
STAKING_AMOUNT="1000000000CELI"
NODEIP="--node http://127.0.0.1:26657"

./celiniumd tendermint unsafe-reset-all
./celiniumd init $VALIDATOR_NAME --chain-id $CHAIN_ID

./celiniumd keys add $KEY_NAME --keyring-backend test
./celiniumd add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test
./celiniumd gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test
./celiniumd collect-gentxs

./celiniumd start   --rpc.laddr tcp://127.0.0.1:46658 --grpc.address 0.0.0.0:10090 

# ./celiniumd tx bank send cosmos1cq6l4pcm50znqg40vl7h69pnzwn90p2fazxatm \
#  cosmos1uva2e4ahemcz40hx3y4j30yasx3g4z2ap6zvel 1000000000CELI --fees 5000CELI --keyring-backend test --chain-id celiniumd-5

# key name ica-control-1 account cosmos17ra6632tf4kz2a9wa7hm8sj6qry3a6w5l83plw

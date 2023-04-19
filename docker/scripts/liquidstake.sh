#!/bin/bash

set -o errexit -o nounset

WALLET_KEY_NAME=$VALIDATOR_NAME 

_register_source_chain(){
    source_chain_id=$2
    connection_id=$3
    transfer_channel_id=$4
    val_prefix=$5
    vals=$6
    native_denom=$7
    derivative_denom=$8

    $CHAIN_NODE tx liquidstake register-source-chain \
        $source_chain_id $connection_id $transfer_channel_id \
        $val_prefix "$vals" $native_denom $derivative_denom\
        --chain-id=$CHAIN_ID \
        --gas="auto" \
        --gas-adjustment=1.5 \
        --fees="5000$DENOM" \
        --from=$WALLET_KEY_NAME \
        --keyring-backend=test
}

_delegate(){
    chain_id=$2
    amount=$3
    from=$4 # from wallet name
    
    $CHAIN_NODE tx liquidstake delegate $chain_id $amount \
        --chain-id=$CHAIN_ID \
        --gas="auto" \
        --gas-adjustment=1.5 \
        --fees="5000$DENOM" \
        --from=$from \
        --keyring-backend=test
}

_undelegate(){
    chain_id=$2
    amount=$3
    from=$4 # from wallet name
    
    $CHAIN_NODE tx liquidstake undelegate $chain_id $amount \
        --chain-id=$CHAIN_ID \
        --gas="auto" \
        --gas-adjustment=1.5 \
        --fees="5000$DENOM" \
        --from=$from \
        --keyring-backend=test
}

_claim(){
    chain_id=$2
    epoch=$3
    from=$4 # from wallet name
    
    $CHAIN_NODE tx liquidstake claim $chain_id $epoch \
        --chain-id=$CHAIN_ID \
        --gas="auto" \
        --gas-adjustment=1.5 \
        --fees="5000$DENOM" \
        --from=$from \
        --keyring-backend=test
}

if [ "$1" = 'register_source_chain' ]; then
    _register_source_chain "$@"
elif [ "$1" = 'delegate' ]; then
   _delegate "$@"
elif [ "$1" = 'undelegate' ]; then
   _undelegate "$@"
elif [ "$1" = 'claim' ]; then
   _claim "$@"
fi
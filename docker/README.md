# Docker Compose Testnet
At this stage, we will use Docker Compose to start a simple local testnet to test basic functionalities. The network will consist of a Gaia (Cosmos Hub) validator node, a Celinium validator node, and a Relayer program.
The Relayer will establish IBC links and channels between the Gaia chain and the Celinium chain, enabling the network to use IBC-transfer and ICS-27.

# Start
```
Step 1: Go to the root directory of the project and build the celinium image
    docker build -t celinium -f Dockerfile .
Step 2: Start the network
    docker compose up
    After executing the command, the Relayer container will automatically run the initialization script and establish IBC links.
    This process may take about 20 blocks. When you see the following message in the relayer logs, 
    it indicates that the channel handshake is successful:
    docker-relayer-1 | ts=2023-04-20T03:18:45.408841Z lvl=info msg="Found termination condition for channel handshake" 
    path_name=gaia-celi-path chain_id=celinium client_id=07-tendermint-0
```
You can use the clear.sh script to quickly clean up the volumes and containers created by docker compose up.

# Test
```
Step 1: Query the wallet address in the celinium container
    docker compose exec celinium /opt/helper.sh wallet:address
    Response: celi1gpsstdwwwzyeau7mc8q9a2vp97qu3prte46f7w
Step 2: Query the wallet balance in the celinium container
    docker compose exec celinium /opt/helper.sh wallet:balance
    Response: 
        balances:
        - amount: "9999999999999999000000000"
          denom: CELI
        pagination:
          next_key: null
          total: "0"
        
Step 3: Transfer from gaia to celinium            
    docker compose exec gaia-validator-1 /opt/helper.sh wallet:ibc_transfer celi12cnq5gy6vgcq93hxqlzqwpcdcvdca6nfycw3px 1000ATOM
    Query the balance again
        balances:
        - amount: "9999999999999999000000000"
          denom: CELI
        - amount: "1000"
          denom: ibc/04C1A8B4EC211C89630916F8424F16DC9611148A5F300C122464CE8E996AABD0
        pagination:
          next_key: null
          total: "0"
    04C1A8B4EC211C89630916F8424F16DC9611148A5F300C122464CE8E996AABD0=Hash(transfer/channel-0/ATOM)

Step 4: Register source chain on celinium                  
    docker compose exec celinium /opt/liquidstake.sh register_source_chain gaia connection-0 channel-0 cosmosvaloper '{"Vals": [{"weight": 100000000,"address":"cosmosvaloper1lgj6z9ujsv2pszwctcem47x8t0ys3tcmvsszte"}]}' ATOM vpATOM

    {"Vals": [{"weight": 100000000,"address":"cosmosvaloper1lgj6z9ujsv2pszwctcem47x8t0ys3tcmvsszte"}]}, 
    this is the target validator for delegation on gaia,you need to search for the "validator_address" field in the 
    ./gaia_validator_1/config/genesis.json file and use its value for subsequent replacement operations.
    Currently, Gaia chain has only one validator.

step5: delegate
    docker compose exec celinium /opt/liquidstake.sh delegate gaia 500 TCELINIUM

step5: undelegate
    docker compose exec celinium /opt/liquidstake.sh undelegate gaia 250 TCELINIUM

step5: claim
    docker compose exec celinium /opt/liquidstake.sh claim gaia 1 TCELINIUM
    
    Where 1 represents the unbonding epoch when the undelegate occurred.    
```
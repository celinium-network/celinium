# Plan
## 最终测试网：
    3 个 Gaia validator 
    1 个 Celinium validator
    1 个 Relayer 
 
## 本地可行测试
    step0：
        1 个 Gaia validator  
        1 个 Celinium validator
        设置 Gaia 的 staking 参数：
        unbonding period: 48 小时?


    step1: 创建两个账户 并转账
            gaia-user：
            celi-user：

    step2: 启动relayer，建立 IBC 通道，开始 updateclient

    step3: 设置 DelegateEpoch(1小时)/UndelegateEpoch(24小时)/ReInvestEpoch（2小时）

    step4: register sourcechain

    step5: delegate

    step6: undelegate

    step7: claim

# 如何启动relayer
```
    参考 https://github.com/cosmos/relayer#Basic-Usage---Relaying-Packets-Across-Chains

    1. rly config init

    2. 手动配置链
    gaia:
        type: cosmos
        value:
            key-directory: /home/manjaro/.relayer/keys/gaia
            key: gaia-0
            chain-id: gaia
            rpc-addr: http://127.0.0.1:46659
            account-prefix: cosmos
            keyring-backend: test
            gas-adjustment: 1.2
            gas-prices: 0.1atom
            min-gas-amount: 0
            debug: true
            timeout: 100s
            block-timeout: ""
            output-format: json
            sign-mode: direct
            extra-codecs: []
            coin-type: 0
            signing-algorithm: ""
            broadcast-mode: batch
            min-loop-duration: 0s   
    3. 为每个链创建账户
        rly keys add gaia gaia-0

    4. 启动链

    5. 给创建的账户转账
         ./celiniumd tx bank send celi1hqu6s5lkr370g0mcx4tg2n037n5njpzsf722ln celi1c35lchkcnupx3etfcjr2h9pqyfrs4hen7rfdkp 1000000000CELI --fees 5000CELI --keyring-backend test --chain-id celinium

    6. 配置path
        rly paths new gaia celinium gaia-celi-path

    7. 创建path
        rly transact link gaia-celi-path

    8. 启动path
        rly start gaia-celi-path       
```

# Command
```
// ibc transfer
gaiad tx ibc-transfer transfer transfer channel-0 celi1hqu6s5lkr370g0mcx4tg2n037n5njpzsf722ln 10000atom --fees 5000atom --keyring-backend test --chain-id gaia --from cosmos1p2p6y9tn7sptamkrh79mlwf484r9y5lwc7j0se

快速清理docker-compose 创建的所有卷/镜像/容器 docker-compose down --volumes --remove-orphans

    docker build -t celinium -f Dockerfile .
```
docker compose exec celinium /opt/liquidstake.sh register_source_chain gaia connection-0 transfer cosmosvaloper '{"Vals": [{"weight": 100000000,"address":"cosmosvaloper1cgkvd2h6xun7s4xrvre3me9psdrktxd4dj494z"}]}' ATOM vpATOM

docker compose exec celinium /opt/helper.sh wallet:balance

docker compose exec celinium /opt/helper.sh wallet:address

docker compose exec gaia-validator-1 /opt/helper.sh wallet:ibc_transfer celi1mt6dvlc777fencqyvd3n22pl3qpr4gcm8ryp68 1000ATOM

# docker
构建镜像时，
# Docker Compose Testnet
在当前阶段，我们将通过 Docker Compose 启动一个简单的本地测试网来测试基本的功能。网络有一个 Gaia（cosmos hub） 验证节点，一个 Celinium 验证节点，一个 Relayer 程序。
Relayer 将在 Gaia 链和 Celinium 之间建立 IBC 链接和通道，使网络可以使用 IBC-transfer和 ICS-27.

# Start
```
step1: 回到项目根目录，构建 celinium 镜像
    docker build -t celinium -f Dockerfile .
step2: 启动网络
    docker compose up
    执行命令后，Relayer 容器会自动执行初始化脚本，建立IBC链接。这个过程大约需要20个块。在观测到relayer日志有以下消息说明通道建立成功，
    docker-relayer-1 | ts=2023-04-20T03:18:45.408841Z lvl=info msg="Found termination condition for channel handshake" 
    path_name=gaia-celi-path chain_id=celinium client_id=07-tendermint-0
   
```
可以通过 clear.sh 脚快速清理 docker compose up 当中创建的卷和容器。

# Test
```
step1: 查询当前 celinium 容器当中钱包地址
    docker compose exec celinium /opt/helper.sh wallet:address
    response: celi1gpsstdwwwzyeau7mc8q9a2vp97qu3prte46f7w
step2: 查询当前 celinium 容器当中钱包资产
    docker compose exec celinium /opt/helper.sh wallet:balance
    response: 
        balances:
        - amount: "9999999999999999000000000"
          denom: CELI
        pagination:
          next_key: null
          total: "0"
        
step3: 从 gaia 转账到 celinium            
    docker compose exec gaia-validator-1 /opt/helper.sh wallet:ibc_transfer celi12cnq5gy6vgcq93hxqlzqwpcdcvdca6nfycw3px 1000ATOM
    再查询资产
        balances:
        - amount: "9999999999999999000000000"
          denom: CELI
        - amount: "1000"
          denom: ibc/04C1A8B4EC211C89630916F8424F16DC9611148A5F300C122464CE8E996AABD0
        pagination:
          next_key: null
          total: "0"
    04C1A8B4EC211C89630916F8424F16DC9611148A5F300C122464CE8E996AABD0=Hash(transfer/channel-0/ATOM)

step4: 在 celinium 上注册 sourcechain，                  
    docker compose exec celinium /opt/liquidstake.sh register_source_chain gaia connection-0 channel-0 cosmosvaloper '{"Vals": [{"weight": 100000000,"address":"cosmosvaloper1lgj6z9ujsv2pszwctcem47x8t0ys3tcmvsszte"}]}' ATOM vpATOM

    {"Vals": [{"weight": 100000000,"address":"cosmosvaloper1lgj6z9ujsv2pszwctcem47x8t0ys3tcmvsszte"}]}， 这是将要在gaia上进行delegate的目标validator，
    这个地址暂时从 ./gaia_validator_1/config/genesis.json 当中搜索“validator_address”，获得的值进行替代，目前 gaia 只有一个Validator。

step5: delegate
    docker compose exec celinium /opt/liquidstake.sh delegate gaia 500 TCELINIUM

step5: undelegate
    docker compose exec celinium /opt/liquidstake.sh undelegate gaia 250 TCELINIUM

step5: claim
    undelegate操作后，再经过 unbond 周期，现在用户可以 claim
    docker compose exec celinium /opt/liquidstake.sh claim gaia 1 TCELINIUM
    1 代表的是undelegate 发生的 unbonding epoch 为 1。         
```
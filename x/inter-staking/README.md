# `x/inter-staking`

## Abstract  

The x/inter-staking objective is to share security through a restaking mechanism.
When launching a new Proof-of-Stake (POS) chain, one typically faces the challenge
of token concentration in the initial stages, leading to network centralization. 
One solution to this problem is to reuse a POS system of a well-functioning chain. 
``` 
                        +-----------------+                        +-----------------+
                        |    POS Chain A  |                        |    POS Chain B  |
                        +-----------------+                        +-----------------+
                                  ^                                       ^
                                  |                                       |
 1. Cross chain Transfer tokenA   |                                       |
                                  |                                       |
                                  |   2. Receive TokenA at ChainB         |
                                  |                                       |
                                  |   3. Transfer TokenA to ChainB's      |
                                  |     interchain account at ChainB -----|->(hook actions: transfer )     
                                  |                                       |
                                  |   4. Mint BTokenA (1:1 ratio)         |
                                  |      to UserA account at ChainB       |
                                  |                                       |
                                  |   5. ChainB's interchain account      |
                                  |      participate ChainA‘s staking     |
                                  |                                       |
                                  |   6. User participate ChainB's        |
                                  |      staking with BTotkenA     -------|-> (hook actions)
                                  |                                       |
                                  |   7. User burn BTotkenA  -------------|-> (hook actions)
                                  |                                       |   
 8. ChainB's interchain account   |                                       |
  unbound from ChainA's staking   |                                       |
                                  |                                       |
 9. After unbound period,ChainB's |                                       |
interchain account get TokenA     |                                       |
 from ChainA's staking            |                                       |  
                                  |                                       |
                                  |                                       |
                                  |  10. ChainB's interchain account      |
                                  |      transfer TokenA and ChainA's     |
                                  |      staking reward to User           |
                                  |                                       |
                                  v                                       v
```
## State
### Source Chain 
The set of source chain validators that the interchain account will delegate to. They are determined by on-chain governance.
This governance process：
```
nominate source chain -> nominate source validators -> Wait for a certain amount of delegate amount -> start
```
The metadata info about source chain.
* SourceChainMetadata `0x11 | SourceChainIdLen(1 byte )| SourceChainId -> ProtocolBuffer(SourceChainMetadata) `

The information of the delegation of the source chain.
* SourceChainDelegation `0x12 | SourceChainIdLen(1 byte )| SourceChainId -> ProtocolBuffer(ICADelegation)`

### Delegation
In a block, the chain may receive multiple delegation/undelegation transactions, and these transactions are all to be called across the chain. Because these delegations are handled uniformly through Chain’s inter-chain accounts. So we can store it first, and then process them after merging at `EndBlock`.
* Delegation: `0x21 | DelegatorAddrLen(1 byte) | DelegatorAddr | SourceChainIdLen(1 byte )| SourceChainId -> ProtocolBuffer(UserDelegation)`
* DelegationTaskQueue `0x22 | blockHeight -> ProtocolBuffer([]DelegationTask)`
* UndelegationTaskQueue `0x23 |blockHeight -> ProtocolBuffer([]UnDelegationTask)` 

### Distribution
Each cross-chain undelegation will cause the inter-chain account to automatically receive the staking and rewards tokens after the end of the source chain unbound period. Then at this time, these tokens need to be distributed to the batch of users who trigger the undelegation.
In `EndBlock`, take out the Undelegation that has reached the time requirement, and distribute the tokens.

* DestributeQueue `0x31 | format(time) | SourceChainIdLen(1 byte )| SourceChainId | DelegatorAddrLen(1 byte) | DelegatorAddr | -> nil`

## State Transaction

### Source Chain
The source chain is added and removed through governance. If there is still related delegation/undelegation, delete will be prohibited.
Every delegate/undelegate will change `SourceChainDelegation`.

At the same time, the staking strategy on the source chain can be changed through governance. Including the selection of the source chain validator to obtain the proportion of user staking.

### Delegate
When the user initiates the delegation, his assets are on the upper Control at this time, the whole delegation process is like this:
```
a. Transfer source chain native token to ICA, then a delegation task push into pending task queue.
b. ICA begin transfer source chain native token back, then move the delegation task in preparing task queue.
c. Control chain receive transfer successful signal, move the delegation to preparing queue.
d. ICA begin staking in source chain, then move then delegation task from preparing queue to ongoing queue.
e. Control chain receive staking successful signal, move the delegation from ongoingqueue, store delegator's info.
```

### Undelegate

### Distribution
Only after the source chain completes undelegate, the control chain can get back the principal and rewards. Obviously the control chain itself cannot receive this signal. Then the relayer must be used to pass some information and proofs to complete the process.
With `IBC` and `cosmos sdk staking module`, the process is as follows
```
a. Control chain receive undelegation task, store it in the queue.
b. Send undelegation to source chain with control chain interchain account. 
c. Source chain insert key-value like { `UnbondingID`: `UnbondingDelegation`} in store.  
    So We can get it and its merkle proof. And at the same time control chain get the active validators 
    information in the IBC block header. Each Validator has an associated `UnbondingID` array.
d. After some time, The relayer observes that the `UnbondingID` is consumed. This means that the `Undelegation`  
    initiated by the control chain has taken effect. The relayer will submit `UnbondingDelegation` and proofs to the control chain.
e. At this time, the control chain detects that `UnbondingID` in the relevant validator in the IBC header no longer exists.
   Then, The control chain starts executing the distribution.   
```

## Messages

### MsgAddSourceChain
The `MsgAddSourceChain` add a new source chain for restaking.
The source chain are updated through a governance proposal where the signer is the gov module account address.

```protobuf reference

```
### MsgUpdateSourceChainDelegatePlan
The `MsgAddSourceChain` update source chain delegation plan.
The staking plan are updated through a governance proposal where the signer is the gov module account address.
```protobuf reference

```

### MsgDelegate
The `MsgDelegate` send delegation message from control chain to source chain.
```protobuf reference

```
### MsgUndelegate
The `MsgDelegate` send undelegation message from control chain to source chain.
```protobuf reference

```

## End Block
### Delegate/Undelegate
Summarize a certain amount of delegation/undelegation from delegation/undelegation task queue，then send message to source chain.               

## Hooks

## Events

## Parameters

## Client
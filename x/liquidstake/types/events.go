package types

// liquidstake module event types
const (
	EventTypeRegisterSourceChain = "register_source_chain"
	EventTypeDelegate            = "delegate"
	EventTypeUndelegate          = "undelegate"

	AttributeKeySourceChainID = "source_chain_id"
	AttributeKeyDelegator     = "delegator"
	AttributeKeyValidators    = "source_chain_validators"
	AttributeKeyWeights       = "source_chain_validator_weights"
	AttributeKeyEpoch         = "epcoh_number"
	AttributeKeyDelegateAmt   = "delegate_amount"
	AttributeKeyRedeemAmt     = "redeem_amount"
	AttributeKeyUnbondAmt     = "unbond_amount"
	AttributeKeyClaimAmt      = "unbond_amount"
)

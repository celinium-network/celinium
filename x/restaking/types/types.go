package types

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

type (
	TendermintABCIValidatorUpdate = abci.ValidatorUpdate
	ValidatorSet                  = tmtypes.ValidatorSet
)

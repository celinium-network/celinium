package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
)

type (
	Int                        = sdkmath.Int
	Dec                        = sdk.Dec
	TendermintLightClientState = ibctmtypes.ClientState
)

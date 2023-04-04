package types

import (
	epochstypes "celinium/x/epochs/types"
)

// epoch identifier for liquid stake.
// todo! identifier update by gov?
const (
	DelegationEpochIdentifier   = epochstypes.HourEpochID
	UndelegationEpochIdentifier = epochstypes.DayEpochID
)

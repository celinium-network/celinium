package types

import (
	epochstypes "github.com/celinium-netwok/celinium/x/epochs/types"
)

// epoch identifier for liquid stake.
// TODO identifier update by gov?
const (
	DelegationEpochIdentifier   = epochstypes.HourEpochID
	UndelegationEpochIdentifier = epochstypes.DayEpochID
)

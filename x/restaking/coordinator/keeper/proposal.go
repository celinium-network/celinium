package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	"github.com/celinium-network/celinium/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/celinium/x/restaking/types"
)

func (k Keeper) HandleConsumerAdditionProposal(ctx sdk.Context, proposal *types.ConsumerAdditionProposal) error {
	chainID := proposal.ChainId

	if _, found := k.GetConsumerClientID(ctx, proposal.ChainId); found {
		return errorsmod.Wrap(restaking.ErrDuplicateConsumerChain,
			fmt.Sprintf("cannot create client for existent consumer chain: %s", chainID))
	}

	k.SetPendingConsumerAdditionProposal(ctx, proposal)

	return nil
}

func verifyConsumerAdditionProposal(proposal *types.ConsumerAdditionProposal, client *ibctmtypes.ClientState) error {
	return nil
}

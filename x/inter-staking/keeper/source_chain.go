package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"

	"celinium/x/inter-staking/types"
)

// SourceChainAvaiable whether SourceChain is available.
func (k Keeper) SourceChainAvaiable(ctx sdk.Context, connectionID string, icaCtladdr string) bool {
	portID, err := icatypes.NewControllerPortID(icaCtladdr)
	if err != nil {
		return false
	}
	_, found := k.icaControllerKeeper.GetOpenActiveChannel(ctx, connectionID, portID)
	return found
}

func (k Keeper) AddSourceChain(
	ctx sdk.Context,
	strategies []types.DelegationStrategy,
	sourceDenom,
	sourceTraceDenom,
	chainID,
	connectionID,
	channelID,
	version string,
) (string, error) {
	if len(strategies) == 0 {
		return "", sdkerrors.Wrapf(types.ErrMismatchParameter, "the delegate plan should be set")
	}

	if exist, _ := k.SourceChainExist(ctx, chainID); exist {
		return "", sdkerrors.Wrapf(types.ErrSourceChainExist, "source chain: %s already exist", chainID)
	}

	icaCtlAccount := types.GenerateSourceChainControlAccount(ctx, chainID, connectionID)

	k.accountKeeper.NewAccount(ctx, icaCtlAccount)
	k.accountKeeper.SetAccount(ctx, icaCtlAccount)

	icaCtladdr := icaCtlAccount.GetAddress().String()

	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, connectionID, icaCtladdr, version); err != nil {
		return "", err
	}

	sourceChainMetaData := types.SourceChainMetadata{
		IbcClientId:           chainID,
		IbcConnectionId:       connectionID,
		IbcTransferChannelId:  channelID,
		ICAControlAddr:        icaCtladdr,
		SourceChainDenom:      sourceDenom,
		SourceChainTraceDenom: sourceTraceDenom,
		DelegateStrategy:      strategies,
	}

	k.SetSourceChain(ctx, chainID, &sourceChainMetaData)

	return icaCtladdr, nil
}

package keeper

import (
	"celinium/x/liquid-stake/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddSouceChain(ctx sdk.Context, sourceChain *types.SourceChain) error {
	if err := sourceChain.BasicVerify(); err != nil {
		return sdkerrors.Wrapf(types.ErrSourceChainParameter, "error: %v", err)
	}

	// check source chain ibc client
	if err := k.checkIBCClient(ctx, sourceChain.ChainID); err != nil {
		return err
	}

	// check source chain ibc transfer.
	// todo!: Should consider whether to detect?
	if !k.ibcTransferKeeper.GetSendEnabled(ctx) || !k.ibcTransferKeeper.GetReceiveEnabled(ctx) {
		return types.ErrBannedIBCTransfer
	}

	// check source chain wheather is already existed.
	if _, found := k.GetSourceChain(ctx, sourceChain.ChainID); found {
		return sdkerrors.Wrapf(types.ErrSourceChainExist, "already exist source chain, ID: %s", sourceChain.ChainID)
	}

	accounts := sourceChain.GenerateAndFillAccount(ctx)

	for _, a := range accounts {
		k.accountKeeper.NewAccount(ctx, a)
		k.accountKeeper.SetAccount(ctx, a)
		if err := k.icaCtlKeeper.RegisterInterchainAccount(ctx, sourceChain.ConnectionID, a.GetAddress().String(), ""); err != nil {
			return err
		}
	}

	return nil
}

// chainAvaiable wheather a chain is available. when all interchain account is registered, then it's available
func (k Keeper) SourceChainAvaiable(ctx sdk.Context, sourceChain *types.SourceChain) bool {
	_, found1 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.WithdrawAddress)
	_, found2 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.DelegateAddress)
	_, found3 := k.icaCtlKeeper.GetInterchainAccountAddress(ctx, sourceChain.ConnectionID, sourceChain.UnboudAddress)

	if found1 && found2 && found3 {
		return true
	}

	return false
}

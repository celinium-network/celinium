package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"celinium/x/inter-staking/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		getAddSourceChainCmd(),
		getDelegateCmd(),
	)

	return cmd
}

func getAddSourceChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "add source chain",
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signer := clientCtx.GetFromAddress().String()

			chainID := args[0]

			connectionID := args[1]

			stakingDenom := args[2]

			var strgtegy []types.DelegationStrategy

			if err := json.Unmarshal([]byte(args[3]), &strgtegy); err != nil {
				return err
			}

			msg := types.NewMsgAddSourceChain(
				chainID,
				connectionID,
				"",
				stakingDenom,
				strgtegy,
				signer,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "delegate to source chain",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegator := clientCtx.GetFromAddress().String()

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return nil
			}

			msg := types.NewMsgDelegate(
				args[0],
				coin,
				delegator,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

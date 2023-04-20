package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	liquistakeQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the staking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquistakeQueryCmd.AddCommand(
		GetSourceChainCmd(),
		GetDelegationRecordCmd(),
		GetChainUnbondingCmd(),
		GetUserDelegationRecordCmd(),
	)

	return liquistakeQueryCmd
}

func GetSourceChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sourcechain [chain_id]",
		Short: "Query a source chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QuerySourceChainRequest{
				ChainID: args[0],
			}
			res, err := queryClient.SourceChain(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.SourceChain)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetDelegationRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation-record [chain_id] [epoch]",
		Short: "Query delegation record of a source chain from the specific epoch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			epoch, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			params := &types.QueryChainEpochDelegationRecordRequest{
				ChainID: args[0],
				Epoch:   uint64(epoch),
			}
			res, err := queryClient.ChainEpochDelegationRecord(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetChainUnbondingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-unbonding [chain_id] [epoch]",
		Short: "Query unbonding of chain in the specific epoch",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			epoch, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			params := &types.QueryChainEpochUnbondingRequest{
				ChainID: args[0],
				Epoch:   uint64(epoch),
			}
			res, err := queryClient.ChainEpochUnbonding(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetUserDelegationRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-undelegation [chain_id] [user_address]",
		Short: "Query undelegation record of the user for the specific chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryUserUndelegationRecordRequest{
				ChainID: args[0],
				User:    args[1],
			}
			res, err := queryClient.UserUndelegationRecord(cmd.Context(), params)
			if err != nil {
				return err
			}

			clientCtx.PrintProto(res)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

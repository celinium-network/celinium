package cli

import (
	"encoding/json"
	"strconv"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/celinium-netwok/celinium/x/liquidstake/types"
)

func NewTxCmd() *cobra.Command {
	liquidStakeTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquidStakeTxCmd.AddCommand(NewRegisterSourceChainCmd())
	liquidStakeTxCmd.AddCommand(NewDelegateCmd())
	liquidStakeTxCmd.AddCommand(NewUndelegateCmd())
	liquidStakeTxCmd.AddCommand(NewClaimCmd())

	return liquidStakeTxCmd
}

type validators struct {
	Vals []types.Validator
}

func NewRegisterSourceChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: `register-source-chain [chain_id] [connection] [transfer_channel_id] [val_prefix] [validators] " +
			"[native_denom] [derivative_denom].`,
		Short: `register a new source chain for liquid stake.The [validators] should be json like \n 
		"validators": [{
			"weight": 1000000,
			"address": "xxxxxx"
		}]`,
		Args: cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChainID := args[0]
			connectionID := args[1]
			transferChannelID := args[2]
			valAddrPrefix := args[3]

			var vals validators
			err = json.Unmarshal([]byte(args[4]), &vals)
			if err != nil {
				return err
			}

			nativeDenom := args[5]
			derivativeDenom := args[5]

			msg := types.MsgRegisterSourceChain{
				ChainID:                   sourceChainID,
				ConnectionID:              connectionID,
				TrasnferChannelID:         transferChannelID,
				Bech32ValidatorAddrPrefix: valAddrPrefix,
				Validators:                vals.Vals,
				NativeDenom:               nativeDenom,
				DerivativeDenom:           derivativeDenom,
				Caller:                    clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegate [chain_id] [amount]`,
		Short: `delegate to the source chain`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChainID := args[0]
			amt, err := math.ParseUint(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgDelegate{
				ChainID:   sourceChainID,
				Amount:    math.NewIntFromBigInt(amt.BigInt()),
				Delegator: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `undelegate [chain_id] [amount].`,
		Short: `delegate form the source chain`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChainID := args[0]

			amt, err := math.ParseUint(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgUndelegate{
				ChainID:   sourceChainID,
				Amount:    math.NewIntFromBigInt(amt.BigInt()),
				Delegator: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `claim [chain_id] [epoch].`,
		Short: `claim funds from complete undelegate `,
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChainID := args[0]
			epoch, err := strconv.ParseUint(args[1], 10, 0)
			if err != nil {
				return err
			}
			msg := types.MsgClaim{
				ChainId:   sourceChainID,
				Epoch:     epoch,
				Delegator: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

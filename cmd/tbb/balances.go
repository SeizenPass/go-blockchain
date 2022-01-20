package main

import (
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/spf13/cobra"
	"os"
)

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances (list...)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	balancesCmd.AddCommand(balancesListCmd)

	return balancesCmd
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List of all balances.",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer state.Close()

		fmt.Println("Accounts balances:")
		fmt.Println("__________________")
		fmt.Println("")

		for account, balance := range state.Balances {
			fmt.Println(fmt.Sprintf("%s: %d", account, balance))
		}
	},
}

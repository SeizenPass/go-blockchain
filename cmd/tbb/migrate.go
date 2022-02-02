package main

import (
	"context"
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/SeizenPass/go-blockchain/node"
	"github.com/spf13/cobra"
	"os"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use: "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			pendingBlock := node.NewPendingBlock(
				database.Hash{},
				state.NextBlockNumber(),
				[]database.Tx{
					database.NewTx("miras", "amiran", 3, ""),
					database.NewTx("miras", "miras", 700, "reward"),
					database.NewTx("amiran", "beknur", 2000, ""),
					database.NewTx("amiran", "amiran", 100, "reward"),
					database.NewTx("beknur", "miras", 1, ""),
					database.NewTx("miras", "miras", 100, "reward"),
					database.NewTx("miras", "miras", 100, "reward"),
					database.NewTx("miras", "miras", 24700, "reward"),
				},
			)

			_, err = node.Mine(context.Background(), pendingBlock)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}

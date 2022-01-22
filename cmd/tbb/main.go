package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const flagDataDir = "datadir"

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(migrateCmd())
	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(runCmd())

	err := tbbCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command)  {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dir where the DB will/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

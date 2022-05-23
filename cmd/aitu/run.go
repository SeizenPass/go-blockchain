package main

import (
	"context"
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/SeizenPass/go-blockchain/node"
	"github.com/spf13/cobra"
	"os"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the AITU node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			miner, _ := cmd.Flags().GetString(flagMiner)
			sslEmail, _ := cmd.Flags().GetString(flagSSLEmail)
			isSSLDisabled, _ := cmd.Flags().GetBool(flagDisableSSL)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)
			bootstrapIp, _ := cmd.Flags().GetString(flagBootstrapIp)
			bootstrapPort, _ := cmd.Flags().GetUint64(flagBootstrapPort)
			bootstrapAcc, _ := cmd.Flags().GetString(flagBootstrapAcc)

			fmt.Println("Launching AITU node and its HTTP API...")

			bootstrap := node.NewPeerNode(
				//TODO: add our own bootstrap node
				bootstrapIp,
				bootstrapPort,
				true,
				database.NewAccount(bootstrapAcc),
				false,
			)

			if !isSSLDisabled {
				port = node.DefaultHTTPort
			}

			n := node.New(getDataDirFromCmd(cmd), ip, port, database.NewAccount(miner), bootstrap,
				node.DefaultMiningDifficulty)
			err := n.Run(context.Background(), isSSLDisabled, sslEmail)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Bool(flagDisableSSL, false, "should the HTTP API SSL certificate be disabled? (default false)")
	runCmd.Flags().String(flagSSLEmail, "", "node's HTTP SSL certificate email")
	runCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards")
	runCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	runCmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposed HTTP port for communication with peers (configurable if SSL is disabled)")
	runCmd.Flags().String(flagBootstrapIp, node.DefaultBootstrapIp, "default bootstrap server to interconnect peers")
	runCmd.Flags().Uint64(flagBootstrapPort, node.DefaultBootstrapPort, "default bootstrap server port to interconnect peers")
	runCmd.Flags().String(flagBootstrapAcc, node.DefaultBootstrapAcc, "default bootstrap account to interconnect peers")

	return runCmd
}

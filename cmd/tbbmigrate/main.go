package main

import (
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"os"
	"time"
)

func main() {
	cwd, _ := os.Getwd()
	state, err := database.NewStateFromDisk(cwd)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("miras", "amiran", 3, ""),
			database.NewTx("miras", "miras", 700, "reward"),
		},
		)

	_ = state.AddBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("amiran", "beknur", 2000, ""),
			database.NewTx("amiran", "amiran", 100, "reward"),
			database.NewTx("beknur", "miras", 1, ""),
			database.NewTx("miras", "miras", 100, "reward"),
			database.NewTx("miras", "miras", 100, "reward"),
			},
		)

	_ = state.AddBlock(block1)
	_, _ = state.Persist()
}

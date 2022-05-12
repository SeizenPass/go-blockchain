package node

import (
	"context"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/SeizenPass/go-blockchain/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNode_Run(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	n := New(datadir, "127.0.0.1", 8085, database.NewAccount("miras"), PeerNode{})

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNode_Mining(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	n := New(datadir, nInfo.IP, nInfo.Port, database.NewAccount("miras"), nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	go func() {
		time.Sleep(time.Second * miningIntervalSeconds / 3)
		tx := database.NewTx("miras", "amiran", 1, "")
		_ = n.AddPendingTX(tx, nInfo)
	}()

	go func() {
		time.Sleep(time.Second*miningIntervalSeconds + 2)
		tx := database.NewTx("miras", "amiran", 2, "")
		_ = n.AddPendingTX(tx, nInfo)
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("2 pending TX not mined into 2 under 30m")
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	mirasAcc := database.NewAccount("miras")
	amiranAcc := database.NewAccount("amiran")

	n := New(datadir, nInfo.IP, nInfo.Port, amiranAcc, nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	tx1 := database.NewTx("miras", "amiran", 1, "")
	tx2 := database.NewTx("miras", "amiran", 2, "")
	tx2Hash, _ := tx2.Hash()

	//TODO should be corrected to the true block with true nonce
	/*
		Mined new Block '000000b1a1afa8f262badf59a5aef2ee1d35775b6b7320f2dfcc411db4476f4a' using PoW����:
		        Height: '1'
		        Nonce: '4028503425'
		        Created: '1643913265'
		        Miner: 'miras'
		        Parent: '0000000000000000000000000000000000000000000000000000000000000000'
		        Attempt: '120997454'
		        Time: 26m52.629146s
	*/
	validPreMinedPb := NewPendingBlock(database.Hash{}, 0, mirasAcc, []database.Tx{tx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		err := n.AddPendingTX(tx1, nInfo)
		if err != nil {
			t.Fatal(err)
		}

		err = n.AddPendingTX(tx2, nInfo)
		if err != nil {
			t.Fatal(err)
		}
	}()

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatal(err)
		}
		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatal("synced block should have canceled mining")
		}

		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]

		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatal("synced block should have canceled mining of already mined transaction")
		}

		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining again the 1 TX not included in synced block")
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)

		startingMirasBalance := n.state.Balances[mirasAcc]
		startingAmiranBalance := n.state.Balances[amiranAcc]

		<-ctx.Done()

		endMirasBalance := n.state.Balances[mirasAcc]
		endAmiranBalance := n.state.Balances[amiranAcc]

		expectedMirasBalance := startingMirasBalance - tx1.Value - tx2.Value + database.BlockReward
		expectedAmiranBalance := startingAmiranBalance + tx1.Value + tx2.Value + database.BlockReward

		if endMirasBalance != expectedMirasBalance {
			t.Fatalf("Miras expected end balance is %d not %d", expectedMirasBalance, endMirasBalance)
		}

		if endAmiranBalance != expectedAmiranBalance {
			t.Fatalf("Amiran expected end balance is %d not %d", expectedAmiranBalance, endAmiranBalance)
		}

		t.Logf("Starting Miras balance: %d", startingMirasBalance)
		t.Logf("Starting Amiran balance: %d", startingAmiranBalance)
		t.Logf("Ending Miras balance: %d", endMirasBalance)
		t.Logf("Ending Amiran balance: %d", endAmiranBalance)
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("was suppose to mine 2 pending TX into 2 valid blocks under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatal("no pending TXs should be left to mine")
	}
}

func getTestDataDirPath() string {
	return filepath.Join(os.TempDir(), ".tbb_test")
}

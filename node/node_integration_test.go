package node

import (
	"context"
	"encoding/json"
	"github.com/SeizenPass/go-blockchain/database"
	"github.com/SeizenPass/go-blockchain/fs"
	"github.com/SeizenPass/go-blockchain/wallet"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testKsMirasAccount = "0xBd5C714b73Dc08B1D1D0A0eE9626635C939e5576"
const testKsAmiranAccount = "0x933f7d71eB25D3EcB051E3FfBD6415dC3ac30507"
const testKsMirasFile = "test_miras--Bd5C714b73Dc08B1D1D0A0eE9626635C939e5576"
const testKsAmiranFile = "test_amiran--933f7d71eB25D3EcB051E3FfBD6415dC3ac30507"
const testKsAccountsPwd = "security123"

func TestNode_Run(t *testing.T) {
	datadir, err := getTestDataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	err = fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	n := New(datadir, "127.0.0.1", 8085, database.NewAccount(DefaultMiner), PeerNode{})

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx, true, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNode_Mining(t *testing.T) {
	dataDir, miras, amiran, err := setupTestNodeDir(1000000)
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(dataDir)

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		amiran,
		true,
	)

	n := New(dataDir, nInfo.IP, nInfo.Port, miras, nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	go func() {
		time.Sleep(time.Second * miningIntervalSeconds / 3)
		tx := database.NewTx(miras, amiran, 1, 1, "")
		signedTx, err := wallet.SignTxWithKeystoreAccount(tx, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			t.Error(err)
			return
		}
		_ = n.AddPendingTX(signedTx, nInfo)
	}()

	go func() {
		time.Sleep(time.Second*miningIntervalSeconds + 2)
		tx := database.NewTx(miras, amiran, 2, 2, "")
		signedTx, err := wallet.SignTxWithKeystoreAccount(tx, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			t.Error(err)
			return
		}
		_ = n.AddPendingTX(signedTx, nInfo)
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

	_ = n.Run(ctx, true, "")

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("2 pending TX not mined into 2 under 30m")
	}
}

func TestNode_ForgedTx(t *testing.T) {
	dataDir, miras, amiran, err := setupTestNodeDir(1000000)
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(dataDir)

	n := New(dataDir, "127.0.0.1", 8085, miras, PeerNode{})
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*15)
	mirasPeerNode := NewPeerNode("127.0.0.1", 8085, false, miras, true)

	txValue := uint(5)
	txNonce := uint(1)
	tx := database.NewTx(miras, amiran, txValue, txNonce, "")

	validSignedTx, err := wallet.SignTxWithKeystoreAccount(tx, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}

	_ = n.AddPendingTX(validSignedTx, mirasPeerNode)

	go func() {
		ticker := time.NewTicker(time.Second * (miningIntervalSeconds - 3))
		wasForgedTxAdded := false

		for {
			select {
			case <-ticker.C:
				if !n.state.LatestBlockHash().IsEmpty() {
					if wasForgedTxAdded && !n.isMining {
						closeNode()
						return
					}

					if !wasForgedTxAdded {
						forgedTx := database.NewTx(miras, amiran, txValue, txNonce, "")
						forgedSignedTx := database.NewSignedTx(forgedTx, validSignedTx.Sig)

						_ = n.AddPendingTX(forgedSignedTx, mirasPeerNode)
						wasForgedTxAdded = true

						time.Sleep(time.Second * (miningIntervalSeconds + 3))
					}
				}
			}
		}
	}()

	_ = n.Run(ctx, true, "")

	if n.state.LatestBlock().Header.Number != 0 {
		t.Fatal("was suppose to mine only one TX. The second TX was forged")
	}

	if n.state.Balances[amiran] != txValue {
		t.Fatal("forged tx succeeded")
	}
}

func TestNode_ReplayedTx(t *testing.T) {
	dataDir, miras, amiran, err := setupTestNodeDir(1000000)
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(dataDir)

	n := New(dataDir, "127.0.0.1", 8085, miras, PeerNode{})
	ctx, closeNode := context.WithCancel(context.Background())
	mirasPeerNode := NewPeerNode("127.0.0.1", 8085, false, miras, true)
	amiranPeerNode := NewPeerNode("127.0.0.1", 8086, false, amiran, true)

	txValue := uint(5)
	txNonce := uint(1)
	tx := database.NewTx(miras, amiran, txValue, txNonce, "")

	signedTx, err := wallet.SignTxWithKeystoreAccount(tx, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}

	_ = n.AddPendingTX(signedTx, mirasPeerNode)

	go func() {
		ticker := time.NewTicker(time.Second * (miningIntervalSeconds - 3))
		wasReplayedTxAdded := false

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlockHash().IsEmpty() {
					if wasReplayedTxAdded && !n.isMining {
						closeNode()
						return
					}

					if !wasReplayedTxAdded {
						n.archivedTXs = make(map[string]database.SignedTx)

						_ = n.AddPendingTX(signedTx, amiranPeerNode)
						wasReplayedTxAdded = true

						time.Sleep(time.Second * (miningIntervalSeconds + 3))
					}
				}
			}
		}
	}()

	_ = n.Run(ctx, true, "")

	if n.state.Balances[amiran] == tx.Value*2 {
		t.Errorf("replayed attack was successful :(")
		return
	}

	if n.state.Balances[amiran] != txValue {
		t.Errorf("replayed attack was successful :(")
		return
	}

	if n.state.LatestBlock().Header.Number == 1 {
		t.Errorf("the second block was not suppose to be persisted because it contained a malicious TX")
		return
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	dataDir, miras, amiran, err := setupTestNodeDir(1000000)
	if err != nil {
		t.Error(err)
	}
	defer fs.RemoveDir(dataDir)

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		amiran,
		true,
	)

	n := New(dataDir, nInfo.IP, nInfo.Port, amiran, nInfo)
	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	tx1 := database.NewTx(miras, amiran, 1, 1, "")
	tx2 := database.NewTx(miras, amiran, 2, 2, "")

	signedTx1, err := wallet.SignTxWithKeystoreAccount(tx1, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}

	signedTx2, err := wallet.SignTxWithKeystoreAccount(tx2, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
	if err != nil {
		t.Error(err)
		return
	}

	tx2Hash, err := signedTx2.Hash()
	if err != nil {
		t.Error(err)
		return
	}

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
	validPreMinedPb := NewPendingBlock(database.Hash{}, 0, miras, []database.SignedTx{signedTx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		err := n.AddPendingTX(signedTx1, nInfo)
		if err != nil {
			t.Fatal(err)
		}

		err = n.AddPendingTX(signedTx2, nInfo)
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

		startingMirasBalance := n.state.Balances[miras]
		startingAmiranBalance := n.state.Balances[amiran]

		<-ctx.Done()

		endMirasBalance := n.state.Balances[miras]
		endAmiranBalance := n.state.Balances[amiran]

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

	_ = n.Run(ctx, true, "")

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("was suppose to mine 2 pending TX into 2 valid blocks under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatal("no pending TXs should be left to mine")
	}
}

func TestNode_MiningSpamTransactions(t *testing.T) {
	mirasBalance := uint(1000)
	amiranBalance := uint(0)
	minerBalance := uint(0)
	minerKey, err := wallet.NewRandomKey()
	if err != nil {
		t.Fatal(err)
	}
	miner := minerKey.Address
	dataDir, miras, amiran, err := setupTestNodeDir(mirasBalance)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.RemoveDir(dataDir)

	n := New(dataDir, "127.0.0.1", 8085, miner, PeerNode{})
	ctx, closeNode := context.WithCancel(context.Background())
	minerPeerNode := NewPeerNode("127.0.0.1", 8085, false, miner, true)

	txValue := uint(200)

	txCount := uint(4)
	for i := uint(1); i <= txCount; i++ {
		time.Sleep(time.Second)

		txNonce := i
		tx := database.NewTx(miras, amiran, txValue, txNonce, "")

		signedTx, err := wallet.SignTxWithKeystoreAccount(tx, miras, testKsAccountsPwd, wallet.GetKeystoreDirPath(dataDir))
		if err != nil {
			t.Fatal(err)
		}

		_ = n.AddPendingTX(signedTx, minerPeerNode)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if !n.state.LatestBlockHash().IsEmpty() {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx, true, "")

	expectedMirasBalance := mirasBalance - (txCount * txValue) - (txCount * database.TxFee)
	expectedAmiranBalance := amiranBalance + (txCount * txValue)
	expectedMinerBalance := minerBalance + database.BlockReward + (txCount * database.TxFee)

	if n.state.Balances[miras] != expectedMirasBalance {
		t.Errorf("Miras balance is incorrect. Expected: %d. Got: %d", expectedMirasBalance, n.state.Balances[miras])
		return
	}

	if n.state.Balances[amiran] != expectedAmiranBalance {
		t.Errorf("Amiran balance is incorrect. Expected: %d. Got: %d", expectedAmiranBalance, n.state.Balances[amiran])
		return
	}

	if n.state.Balances[miner] != expectedMinerBalance {
		t.Errorf("Miner balance is incorrect. Expected: %d. Got: %d", expectedMinerBalance, n.state.Balances[miner])
	}

	t.Logf("Miras final balance: %d AITU", n.state.Balances[miras])
	t.Logf("Amiran final balance: %d AITU", n.state.Balances[amiran])
	t.Logf("Miner final balance: %d AITU", n.state.Balances[miner])
}

func getTestDataDirPath() (string, error) {
	return ioutil.TempDir(os.TempDir(), "aitu_test")
}

func copyKeystoreFilesIntoTestDataDirPath(dataDir string) error {
	mirasSrcKs, err := os.Open(testKsMirasFile)
	if err != nil {
		return err
	}
	defer mirasSrcKs.Close()

	ksDir := filepath.Join(wallet.GetKeystoreDirPath(dataDir))

	err = os.Mkdir(ksDir, 0777)
	if err != nil {
		return err
	}

	mirasDstKs, err := os.Create(filepath.Join(ksDir, testKsMirasFile))
	if err != nil {
		return err
	}
	defer mirasDstKs.Close()

	_, err = io.Copy(mirasDstKs, mirasSrcKs)
	if err != nil {
		return err
	}

	amiranSrcKs, err := os.Open(testKsAmiranFile)
	if err != nil {
		return err
	}
	defer amiranSrcKs.Close()

	amiranDstKs, err := os.Create(filepath.Join(ksDir, testKsAmiranFile))
	if err != nil {
		return err
	}
	defer amiranDstKs.Close()

	_, err = io.Copy(amiranDstKs, amiranSrcKs)
	if err != nil {
		return err
	}

	return nil
}

func setupTestNodeDir(mirasBalance uint) (dataDir string, miras, amiran common.Address, err error) {
	miras = database.NewAccount(testKsMirasAccount)
	amiran = database.NewAccount(testKsAmiranAccount)

	dataDir, err = getTestDataDirPath()
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	genesisBalances := make(map[common.Address]uint)
	genesisBalances[miras] = mirasBalance
	genesis := database.Genesis{Balances: genesisBalances}
	genesisJson, err := json.Marshal(genesis)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	err = database.InitDataDirIfNotExists(dataDir, genesisJson)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	err = copyKeystoreFilesIntoTestDataDirPath(dataDir)
	if err != nil {
		return "", common.Address{}, common.Address{}, err
	}

	return dataDir, miras, amiran, nil
}

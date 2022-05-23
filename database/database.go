package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

func GetBlocksAfter(blockHash Hash, dataDir string) ([]Block, error) {
	f, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}

	blocks := make([]Block, 0)
	shouldStartCollecting := false

	if reflect.DeepEqual(blockHash, Hash{}) {
		shouldStartCollecting = true
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFs BlockFS
		err = json.Unmarshal(scanner.Bytes(), &blockFs)
		if err != nil {
			return nil, err
		}

		if shouldStartCollecting {
			blocks = append(blocks, blockFs.Value)
			continue
		}

		if blockHash == blockFs.Key {
			shouldStartCollecting = true
		}
	}

	return blocks, nil
}

func GetBlockByHeightOrHash(state *State, height uint64, hash, dataDir string) (BlockFS, error) {
	var block BlockFS

	key, ok := state.HeightCache[height]
	if hash != "" {
		key, ok = state.HashCache[hash]
	}

	if !ok {
		if hash != "" {
			return block, fmt.Errorf("invalid hash: '%v'", hash)
		}
		return block, fmt.Errorf("invalid height: '%v'", height)
	}

	f, err := os.OpenFile(getBlocksDbFilePath(dataDir), os.O_RDONLY, 0600)
	if err != nil {
		return block, err
	}
	defer f.Close()

	_, err = f.Seek(key, 0)
	if err != nil {
		return block, err
	}
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return block, err
		}
		err = json.Unmarshal(scanner.Bytes(), &block)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return block, err
		}
	}

	return block, nil
}

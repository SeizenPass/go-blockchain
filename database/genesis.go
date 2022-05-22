package database

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
)

var genesisJson = `
	{
	  "genesis_time": "2022-01-01T00:00:00.000000000Z",
	  "chain_id": "go-blockchain-ledger",
	  "balances": {
		"0x77a13F4cf2cE723f0794F37eeC7635Dd65AE2736": 1000000,
		"0x9Fd598035Ec0DD0909c054E855272b56F2BeC5C8": 1000000,
		"0x058DF0c85de392cc5bef6c749FE6DD8881a2CA44": 1000000
	  },
	  "fork_aip_1": 5
	}
`

type Genesis struct {
	Balances map[common.Address]uint `json:"balances"`

	ForkAIP1 uint64 `json:"fork_aip_1"`
}

func loadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string, genesis []byte) error {
	return ioutil.WriteFile(path, genesis, 0644)
}

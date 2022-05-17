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
	  }
	}
`

type genesis struct {
	Balances map[common.Address]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return ioutil.WriteFile(path, []byte(genesisJson), 0644)
}

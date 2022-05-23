package node

import (
	"encoding/json"
	"fmt"
	"github.com/SeizenPass/go-blockchain/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBlockExplorer(t *testing.T) {
	tc := []struct {
		arg  string
		want uint64
	}{
		{"1", 1},
		{"3", 3},
		{"99", 99},
	}
	datadir := "test_block_explorer_db"

	n := New(datadir, "127.0.0.1", 8085, database.NewAccount(DefaultMiner), PeerNode{}, 3)

	t.Log(fmt.Sprintf("Listening on: %s:%d", n.info.IP, n.info.Port))

	state, err := database.NewStateFromDisk(n.dataDir, n.miningDifficulty)
	if err != nil {
		t.Fatal(err)
	}
	defer state.Close()

	n.state = state

	pendingState := state.Copy()
	n.pendingState = &pendingState

	t.Log("Blockchain state:")
	t.Logf("			- height: %d\n", n.state.LatestBlock().Header.Number)
	t.Logf("			- hash: %s\n", n.state.LatestBlockHash().Hex())

	for _, tc := range tc {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/block/"+tc.arg, nil)

		func(w http.ResponseWriter, r *http.Request, node *Node) {
			blockByNumberOrHash(w, r, node)
		}(rr, req, n)

		if rr.Code != http.StatusOK {
			if tc.want == 99 {
				continue
			}
			t.Error("unexpected status code: ", rr.Code, rr.Body.String())
		}

		resp := new(database.BlockFS)
		dec := json.NewDecoder(rr.Body)
		err = dec.Decode(resp)
		if err != nil {
			t.Error("error decoding", err)
		}

		got := resp.Value.Header.Number
		if got != tc.want {
			t.Errorf("block explorer(%q) = %v; want %v", tc.arg, got, tc.want)
		}
	}
}

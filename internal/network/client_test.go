package network_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
	"toy-blockchain/internal/transaction"

	"encoding/json"
	"net/http"

	"toy-blockchain/block"
)

func TestFetchStatus(t *testing.T) {
	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	peerAPI := api.NewServer(peerNode)

	peerServer := httptest.NewServer(
		peerAPI.Handler(),
	)
	defer peerServer.Close()

	client := network.NewClient()

	status, err := client.FetchStatus(
		context.Background(),
		peerServer.URL,
	)
	if err != nil {
		t.Fatalf(
			"fetch peer status failed: %v",
			err,
		)
	}

	if status.Address != "peer-node" {
		t.Fatalf(
			"expected address peer-node, got %s",
			status.Address,
		)
	}

	if status.Height != 0 {
		t.Fatalf(
			"expected height 0, got %d",
			status.Height,
		)
	}

	if status.HeadHash == "" {
		t.Fatal(
			"expected non-empty peer head hash",
		)
	}
}
func TestSendTransaction(t *testing.T) {
	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	peerAPI := api.NewServer(peerNode)

	peerServer := httptest.NewServer(
		peerAPI.Handler(),
	)
	defer peerServer.Close()

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	client := network.NewClient()

	err := client.SendTransaction(
		context.Background(),
		peerServer.URL,
		tx,
	)

	if err != nil {
		t.Fatalf(
			"send transaction failed: %v",
			err,
		)
	}

	if peerNode.PendingCount() != 1 {
		t.Fatalf(
			"expected peer pending count 1, got %d",
			peerNode.PendingCount(),
		)
	}
}
func TestTransactionPropagatesToPeer(t *testing.T) {
	nodeB := node.New(
		"node-b",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	serverB := httptest.NewServer(
		api.NewServer(nodeB).Handler(),
	)
	defer serverB.Close()

	nodeA := node.New(
		"node-a",
		[]string{
			serverB.URL,
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	serverA := httptest.NewServer(
		api.NewServer(nodeA).Handler(),
	)
	defer serverA.Close()

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	client := network.NewClient()

	if err := client.SendTransaction(
		context.Background(),
		serverA.URL,
		tx,
	); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	if nodeA.PendingCount() != 1 {
		t.Fatalf(
			"expected node A pending count 1, got %d",
			nodeA.PendingCount(),
		)
	}

	if nodeB.PendingCount() != 1 {
		t.Fatalf(
			"expected node B pending count 1, got %d",
			nodeB.PendingCount(),
		)
	}
}
func TestSendBlock(t *testing.T) {
	minedBlock := block.NewBlock(
		1,
		nil,
		"previous-hash",
	)

	minedBlock.MineConcurrent(
		1,
		1,
	)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.Method != http.MethodPost {
					t.Fatalf(
						"expected POST, got %s",
						r.Method,
					)
				}

				if r.URL.Path != "/blocks" {
					t.Fatalf(
						"expected /blocks, got %s",
						r.URL.Path,
					)
				}

				var received block.Block

				if err := json.NewDecoder(
					r.Body,
				).Decode(&received); err != nil {
					t.Fatalf(
						"failed to decode block: %v",
						err,
					)
				}

				if received.Hash != minedBlock.Hash {
					t.Fatalf(
						"expected block hash %s, got %s",
						minedBlock.Hash,
						received.Hash,
					)
				}

				w.WriteHeader(
					http.StatusCreated,
				)
			},
		),
	)
	defer server.Close()

	client := network.NewClient()

	if err := client.SendBlock(
		context.Background(),
		server.URL,
		minedBlock,
	); err != nil {
		t.Fatalf(
			"send block failed: %v",
			err,
		)
	}
}
func TestFetchBlocks(t *testing.T) {
	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := peerNode.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	minedBlock, _, err := peerNode.MinePending()
	if err != nil {
		t.Fatalf(
			"failed to mine block: %v",
			err,
		)
	}

	peerServer := httptest.NewServer(
		api.NewServer(peerNode).Handler(),
	)
	defer peerServer.Close()

	client := network.NewClient()

	blocks, err := client.FetchBlocks(
		context.Background(),
		peerServer.URL,
		1,
	)
	if err != nil {
		t.Fatalf(
			"fetch blocks failed: %v",
			err,
		)
	}

	if len(blocks) != 1 {
		t.Fatalf(
			"expected 1 block, got %d",
			len(blocks),
		)
	}

	if blocks[0].Height != 1 {
		t.Fatalf(
			"expected block height 1, got %d",
			blocks[0].Height,
		)
	}

	if blocks[0].Hash != minedBlock.Hash {
		t.Fatalf(
			"expected hash %s, got %s",
			minedBlock.Hash,
			blocks[0].Hash,
		)
	}
}

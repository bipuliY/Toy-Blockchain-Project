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

package network_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
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

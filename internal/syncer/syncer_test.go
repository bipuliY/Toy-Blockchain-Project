package syncer_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
	"toy-blockchain/internal/syncer"
	"toy-blockchain/internal/transaction"
)

func TestSyncFromPeer(t *testing.T) {
	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx1 := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := peerNode.SubmitTransaction(tx1); err != nil {
		t.Fatalf(
			"submit first transaction failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"mine first block failed: %v",
			err,
		)
	}

	tx2 := transaction.New(
		transaction.Faucet,
		"Bob",
		50,
	)

	if err := peerNode.SubmitTransaction(tx2); err != nil {
		t.Fatalf(
			"submit second transaction failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"mine second block failed: %v",
			err,
		)
	}

	peerServer := httptest.NewServer(
		api.NewServer(peerNode).Handler(),
	)
	defer peerServer.Close()

	freshNode := node.New(
		"fresh-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	client := network.NewClient()

	if err := syncer.SyncFromPeer(
		context.Background(),
		freshNode,
		client,
		peerServer.URL,
	); err != nil {
		t.Fatalf(
			"sync failed: %v",
			err,
		)
	}

	if freshNode.Height() != peerNode.Height() {
		t.Fatalf(
			"expected height %d, got %d",
			peerNode.Height(),
			freshNode.Height(),
		)
	}

	if freshNode.HeadHash() != peerNode.HeadHash() {
		t.Fatalf(
			"expected matching head hashes: peer=%s fresh=%s",
			peerNode.HeadHash(),
			freshNode.HeadHash(),
		)
	}
}

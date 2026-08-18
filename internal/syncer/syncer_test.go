package syncer_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
	"toy-blockchain/internal/syncer"
	"toy-blockchain/internal/transaction"
)

func TestSyncFromPeerDetectsFork(
	t *testing.T,
) {
	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	localNode := node.New(
		"local-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	// Peer mines its own Block 1.
	peerTx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := peerNode.SubmitTransaction(
		peerTx,
	); err != nil {
		t.Fatalf(
			"peer submit failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"peer mining failed: %v",
			err,
		)
	}

	// Local node mines a DIFFERENT Block 1.
	localTx := transaction.New(
		transaction.Faucet,
		"Bob",
		50,
	)

	if err := localNode.SubmitTransaction(
		localTx,
	); err != nil {
		t.Fatalf(
			"local submit failed: %v",
			err,
		)
	}

	if _, _, err := localNode.MinePending(); err != nil {
		t.Fatalf(
			"local mining failed: %v",
			err,
		)
	}

	if peerNode.Height() != localNode.Height() {
		t.Fatal(
			"expected nodes to have same height",
		)
	}

	if peerNode.HeadHash() == localNode.HeadHash() {
		t.Fatal(
			"expected different head hashes",
		)
	}

	peerServer := httptest.NewServer(
		api.NewServer(peerNode).Handler(),
	)
	defer peerServer.Close()

	client := network.NewClient()

	err := syncer.SyncFromPeer(
		context.Background(),
		localNode,
		client,
		peerServer.URL,
	)

	if !errors.Is(
		err,
		syncer.ErrForkDetected,
	) {
		t.Fatalf(
			"expected fork detected error, got %v",
			err,
		)
	}

	// Detection must NOT replace the chain yet.
	if localNode.HeadHash() == peerNode.HeadHash() {
		t.Fatal(
			"fork detection should not replace local chain",
		)
	}
}

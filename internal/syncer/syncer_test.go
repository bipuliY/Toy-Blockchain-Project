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
func TestSyncFromPeerAdoptsLongerFork(
	t *testing.T,
) {
	localNode := node.New(
		"local-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	peerNode := node.New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	// Local creates A1.
	// This transaction should become orphaned later.
	orphanTx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := localNode.SubmitTransaction(
		orphanTx,
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

	oldLocalHead := localNode.HeadHash()

	// Peer creates a DIFFERENT B1.
	peerTx1 := transaction.New(
		transaction.Faucet,
		"Bob",
		50,
	)

	if err := peerNode.SubmitTransaction(
		peerTx1,
	); err != nil {
		t.Fatalf(
			"peer first submit failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"peer first mining failed: %v",
			err,
		)
	}

	// Peer creates B2, making its branch longer.
	peerTx2 := transaction.New(
		transaction.Faucet,
		"Carol",
		25,
	)

	if err := peerNode.SubmitTransaction(
		peerTx2,
	); err != nil {
		t.Fatalf(
			"peer second submit failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"peer second mining failed: %v",
			err,
		)
	}

	peerServer := httptest.NewServer(
		api.NewServer(peerNode).Handler(),
	)
	defer peerServer.Close()

	client := network.NewClient()

	if err := syncer.SyncFromPeer(
		context.Background(),
		localNode,
		client,
		peerServer.URL,
	); err != nil {
		t.Fatalf(
			"sync/reorg failed: %v",
			err,
		)
	}

	// Local should now have adopted the longer peer chain.
	if localNode.Height() != peerNode.Height() {
		t.Fatalf(
			"expected height %d, got %d",
			peerNode.Height(),
			localNode.Height(),
		)
	}

	if localNode.HeadHash() != peerNode.HeadHash() {
		t.Fatalf(
			"expected local head %s, got %s",
			peerNode.HeadHash(),
			localNode.HeadHash(),
		)
	}

	if localNode.HeadHash() == oldLocalHead {
		t.Fatal(
			"expected old local branch to be replaced",
		)
	}

	// Alice's transaction was inside orphaned A1.
	// It should now be back in the mempool.
	if localNode.PendingCount() != 1 {
		t.Fatalf(
			"expected 1 restored orphaned transaction, got %d",
			localNode.PendingCount(),
		)
	}
}

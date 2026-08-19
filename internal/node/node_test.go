package node

import (
	"errors"
	"testing"

	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"toy-blockchain/chain"
	"toy-blockchain/internal/transaction"
)

func TestNewNode(t *testing.T) {
	n := New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	if n.Address() != "localhost:8001" {
		t.Fatalf(
			"expected address localhost:8001, got %s",
			n.Address(),
		)
	}

	if n.Height() != 0 {
		t.Fatalf(
			"expected genesis height 0, got %d",
			n.Height(),
		)
	}

	if n.HeadHash() == "" {
		t.Fatal("expected genesis block head hash")
	}

	if len(n.Peers()) != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			len(n.Peers()),
		)
	}
}
func TestNodeStatus(t *testing.T) {

	n := New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	status := n.Status()

	if status.Address != "localhost:8001" {
		t.Fatalf(
			"expected address localhost:8001, got %s",
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
		t.Fatal("expected genesis head hash")
	}

	if status.PeerCount != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			status.PeerCount,
		)
	}
	if status.PendingCount != 0 {
		t.Fatalf(
			"expected pending count 0, got %d",
			status.PendingCount,
		)
	}
}
func TestSubmitTransactionRejectsUnsignedTransaction(t *testing.T) {
	n := New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		"Alice",
		"Bob",
		10,
	)

	err := n.SubmitTransaction(tx)

	if err == nil {
		t.Fatal(
			"expected unsigned transaction to be rejected",
		)
	}

	if n.PendingCount() != 0 {
		t.Fatalf(
			"expected pending count 0, got %d",
			n.PendingCount(),
		)
	}
}
func TestSubmitTransactionAcceptsSignedTransaction(t *testing.T) {
	n := New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	publicKey, privateKey, err := ed25519.GenerateKey(
		rand.Reader,
	)
	if err != nil {
		t.Fatalf(
			"failed to generate key pair: %v",
			err,
		)
	}

	sender := hex.EncodeToString(publicKey)

	fundingTx := transaction.New(
		transaction.Faucet,
		sender,
		100,
	)

	if err := n.SubmitTransaction(fundingTx); err != nil {
		t.Fatalf(
			"failed to fund sender: %v",
			err,
		)
	}

	tx := transaction.New(
		sender,
		"Bob",
		25,
	)

	if err := tx.Sign(
		hex.EncodeToString(privateKey),
	); err != nil {
		t.Fatalf(
			"failed to sign transaction: %v",
			err,
		)
	}

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"expected signed transaction to be accepted: %v",
			err,
		)
	}

	if n.PendingCount() != 2 {
		t.Fatalf(
			"expected 2 pending transactions, got %d",
			n.PendingCount(),
		)
	}
}
func TestSubmitTransactionRejectsDuplicate(t *testing.T) {
	n := New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"first submission failed: %v",
			err,
		)
	}

	err := n.SubmitTransaction(tx)

	if !errors.Is(
		err,
		ErrTransactionAlreadySeen,
	) {
		t.Fatalf(
			"expected duplicate transaction error, got %v",
			err,
		)
	}

	if n.PendingCount() != 1 {
		t.Fatalf(
			"expected 1 pending transaction, got %d",
			n.PendingCount(),
		)
	}
}
func TestMinePending(t *testing.T) {
	n := New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	if n.PendingCount() != 1 {
		t.Fatalf(
			"expected 1 pending transaction before mining, got %d",
			n.PendingCount(),
		)
	}

	minedBlock, _, err := n.MinePending()
	if err != nil {
		t.Fatalf(
			"failed to mine pending transaction: %v",
			err,
		)
	}

	if minedBlock.Height != 1 {
		t.Fatalf(
			"expected mined block height 1, got %d",
			minedBlock.Height,
		)
	}

	if n.Height() != 1 {
		t.Fatalf(
			"expected node height 1, got %d",
			n.Height(),
		)
	}

	if n.PendingCount() != 0 {
		t.Fatalf(
			"expected pending count 0 after mining, got %d",
			n.PendingCount(),
		)
	}
}
func TestAcceptBlock(t *testing.T) {
	nodeA := New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	nodeB := New(
		"localhost:8002",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	// Simulate transaction gossip.
	if err := nodeA.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"node A submit failed: %v",
			err,
		)
	}

	if err := nodeB.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"node B submit failed: %v",
			err,
		)
	}

	minedBlock, _, err := nodeA.MinePending()
	if err != nil {
		t.Fatalf(
			"node A mining failed: %v",
			err,
		)
	}

	if err := nodeB.AcceptBlock(minedBlock); err != nil {
		t.Fatalf(
			"node B rejected valid block: %v",
			err,
		)
	}

	if nodeB.Height() != 1 {
		t.Fatalf(
			"expected node B height 1, got %d",
			nodeB.Height(),
		)
	}

	if nodeB.HeadHash() != minedBlock.Hash {
		t.Fatalf(
			"expected node B head %s, got %s",
			minedBlock.Hash,
			nodeB.HeadHash(),
		)
	}

	if nodeB.PendingCount() != 0 {
		t.Fatalf(
			"expected node B pending count 0, got %d",
			nodeB.PendingCount(),
		)
	}
}
func TestAcceptBlockRejectsDuplicate(
	t *testing.T,
) {
	nodeA := New(
		"node-a",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	nodeB := New(
		"node-b",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := nodeA.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"submit transaction failed: %v",
			err,
		)
	}

	minedBlock, _, err := nodeA.MinePending()
	if err != nil {
		t.Fatalf(
			"mine block failed: %v",
			err,
		)
	}

	if err := nodeB.AcceptBlock(minedBlock); err != nil {
		t.Fatalf(
			"first block acceptance failed: %v",
			err,
		)
	}

	err = nodeB.AcceptBlock(minedBlock)

	if !errors.Is(
		err,
		ErrBlockAlreadySeen,
	) {
		t.Fatalf(
			"expected duplicate block error, got %v",
			err,
		)
	}

	if nodeB.Height() != 1 {
		t.Fatalf(
			"expected height 1, got %d",
			nodeB.Height(),
		)
	}
}
func TestBlocksFrom(t *testing.T) {
	n := New(
		"node-a",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	minedBlock, _, err := n.MinePending()
	if err != nil {
		t.Fatalf(
			"failed to mine block: %v",
			err,
		)
	}

	blocks, err := n.BlocksFrom(1)
	if err != nil {
		t.Fatalf(
			"BlocksFrom failed: %v",
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
func TestChainSnapshot(t *testing.T) {
	n := New(
		"node-a",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"submit transaction failed: %v",
			err,
		)
	}

	minedBlock, _, err := n.MinePending()
	if err != nil {
		t.Fatalf(
			"mine block failed: %v",
			err,
		)
	}

	blocks := n.ChainSnapshot()

	if len(blocks) != 2 {
		t.Fatalf(
			"expected 2 blocks including genesis, got %d",
			len(blocks),
		)
	}

	if blocks[1].Hash != minedBlock.Hash {
		t.Fatalf(
			"expected mined block hash %s, got %s",
			minedBlock.Hash,
			blocks[1].Hash,
		)
	}

	// Modify the returned copy.
	blocks[1].Transactions[0].To = "Changed"

	// Get another snapshot.
	fresh := n.ChainSnapshot()

	if fresh[1].Transactions[0].To != "Alice" {
		t.Fatal(
			"modifying snapshot changed internal blockchain",
		)
	}
}
func TestValidateCandidateChain(
	t *testing.T,
) {
	peerNode := New(
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
			"submit transaction failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"mine block failed: %v",
			err,
		)
	}

	candidate := peerNode.ChainSnapshot()

	localNode := New(
		"local-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	if err := localNode.ValidateCandidateChain(
		candidate,
	); err != nil {
		t.Fatalf(
			"expected candidate chain to be valid: %v",
			err,
		)
	}

	// Validation must not adopt the chain.
	if localNode.Height() != 0 {
		t.Fatalf(
			"expected local height 0, got %d",
			localNode.Height(),
		)
	}
}
func TestValidateCandidateChainRejectsTampering(
	t *testing.T,
) {
	peerNode := New(
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
			"submit transaction failed: %v",
			err,
		)
	}

	if _, _, err := peerNode.MinePending(); err != nil {
		t.Fatalf(
			"mine block failed: %v",
			err,
		)
	}

	candidate := peerNode.ChainSnapshot()

	// Deliberately corrupt the block.
	candidate[1].Hash = "tampered"

	localNode := New(
		"local-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	if err := localNode.ValidateCandidateChain(
		candidate,
	); err == nil {
		t.Fatal(
			"expected tampered candidate chain to be rejected",
		)
	}
}

func TestAdoptCandidateChain(
	t *testing.T,
) {
	localNode := New(
		"local-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	peerNode := New(
		"peer-node",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	// Local node creates its own Block 1.
	localTx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
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

	oldLocalHead := localNode.HeadHash()

	// Peer creates a different Block 1.
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

	// Peer then creates Block 2,
	// making its chain longer.
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

	candidate := peerNode.ChainSnapshot()

	if err := localNode.AdoptCandidateChain(
		candidate,
	); err != nil {
		t.Fatalf(
			"adopt candidate failed: %v",
			err,
		)
	}

	if localNode.Height() != peerNode.Height() {
		t.Fatalf(
			"expected height %d, got %d",
			peerNode.Height(),
			localNode.Height(),
		)
	}

	if localNode.HeadHash() != peerNode.HeadHash() {
		t.Fatalf(
			"expected adopted head %s, got %s",
			peerNode.HeadHash(),
			localNode.HeadHash(),
		)
	}

	if localNode.HeadHash() == oldLocalHead {
		t.Fatal(
			"expected local chain to be replaced",
		)
	}
}

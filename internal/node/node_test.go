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

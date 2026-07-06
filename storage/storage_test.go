package storage

import (
	"path/filepath"
	"testing"

	"toy-blockchain/chain"
	"toy-blockchain/internal/transaction"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chain.json")

	bc := chain.NewBlockchain(1, 5)
	if err := bc.AddTransaction(transaction.New(transaction.Faucet, "Alice", 100)); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}

	if err := Save(path, bc); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Difficulty != bc.Difficulty {
		t.Fatalf("difficulty mismatch")
	}

	if len(loaded.PendingTransactions) != 1 {
		t.Fatalf("expected one pending transaction")
	}
}

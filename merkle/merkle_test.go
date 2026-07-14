package merkle

import (
	"testing"

	"toy-blockchain/internal/transaction"
)

func TestCalculateRootIsDeterministic(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
		{
			From:   "Alice",
			To:     "Bob",
			Amount: 30,
		},
	}

	first := CalculateRoot(transactions)
	second := CalculateRoot(transactions)

	if first != second {
		t.Fatalf(
			"expected deterministic Merkle root, got %s and %s",
			first,
			second,
		)
	}
}

func TestChangingTransactionChangesRoot(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
		{
			From:   "Alice",
			To:     "Bob",
			Amount: 30,
		},
	}

	originalRoot := CalculateRoot(transactions)

	transactions[1].Amount = 999

	modifiedRoot := CalculateRoot(transactions)

	if originalRoot == modifiedRoot {
		t.Fatal("expected changed transaction to change the Merkle root")
	}
}

func TestTransactionOrderChangesRoot(t *testing.T) {
	firstOrder := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
		{
			From:   transaction.Faucet,
			To:     "Bob",
			Amount: 50,
		},
	}

	secondOrder := []transaction.Transaction{
		firstOrder[1],
		firstOrder[0],
	}

	firstRoot := CalculateRoot(firstOrder)
	secondRoot := CalculateRoot(secondOrder)

	if firstRoot == secondRoot {
		t.Fatal("expected transaction order to affect the Merkle root")
	}
}

func TestCalculateRootHandlesOddTransactionCount(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
		{
			From:   transaction.Faucet,
			To:     "Bob",
			Amount: 50,
		},
		{
			From:   transaction.Faucet,
			To:     "Charlie",
			Amount: 25,
		},
	}

	root := CalculateRoot(transactions)

	if root == "" {
		t.Fatal("expected Merkle root for odd transaction count")
	}

	if len(root) != 64 {
		t.Fatalf(
			"expected SHA-256 root length of 64 characters, got %d",
			len(root),
		)
	}
}

func TestCalculateRootHandlesEmptyTransactions(t *testing.T) {
	first := CalculateRoot(nil)
	second := CalculateRoot([]transaction.Transaction{})

	if first == "" {
		t.Fatal("expected a root for an empty transaction list")
	}

	if first != second {
		t.Fatal("expected deterministic empty transaction root")
	}

	if len(first) != 64 {
		t.Fatalf(
			"expected SHA-256 root length of 64 characters, got %d",
			len(first),
		)
	}
}
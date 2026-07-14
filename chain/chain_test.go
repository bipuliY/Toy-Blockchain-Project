package chain

import (
	"testing"

	"toy-blockchain/internal/transaction"
)

func TestHonestChainValidates(t *testing.T) {
	bc := NewBlockchain(2, 5)

	if err := bc.AddTransaction(transaction.New(transaction.Faucet, "Alice", 100)); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}

	if _, _, err := bc.MinePending(); err != nil {
		t.Fatalf("unexpected mining error: %v", err)
	}

	if err := bc.AddTransaction(transaction.New("Alice", "Bob", 30)); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}

	if _, _, err := bc.MinePending(); err != nil {
		t.Fatalf("unexpected mining error: %v", err)
	}

	result := bc.Validate()
	if !result.Valid {
		t.Fatalf("expected chain to be valid, got: %s", result.Reason)
	}
}

func TestTamperingIsDetected(t *testing.T) {
	bc := NewBlockchain(2, 5)

	if err := bc.AddTransaction(transaction.New(transaction.Faucet, "Alice", 100)); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}

	if _, _, err := bc.MinePending(); err != nil {
		t.Fatalf("unexpected mining error: %v", err)
	}

	bc.Blocks[1].Transactions[0].Amount = 999

	result := bc.Validate()
	if result.Valid {
		t.Fatalf("expected tampered chain to be invalid")
	}

	if result.BlockHeight != 1 {
		t.Fatalf("expected block 1 to be reported, got %d", result.BlockHeight)
	}
}

func TestOverspendingTransactionIsRejected(t *testing.T) {
	bc := NewBlockchain(1, 5)

	if err := bc.AddTransaction(transaction.New(transaction.Faucet, "Alice", 100)); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}

	if _, _, err := bc.MinePending(); err != nil {
		t.Fatalf("unexpected mining error: %v", err)
	}

	err := bc.AddTransaction(transaction.New("Alice", "Bob", 150))
	if err == nil {
		t.Fatalf("expected overspending transaction to be rejected")
	}

	balances := bc.Balances()
	if balances["Alice"] != 100 {
		t.Fatalf("expected Alice balance unchanged at 100, got %d", balances["Alice"])
	}
}

func TestNegativeAmountIsRejected(t *testing.T) {
	bc := NewBlockchain(1, 5)

	err := bc.AddTransaction(transaction.New(transaction.Faucet, "Alice", -10))
	if err == nil {
		t.Fatalf("expected negative amount to be rejected")
	}
}
func TestValidationDetectsMerkleRootMismatch(t *testing.T) {
	blockchain := NewBlockchain(1, 5)

	err := blockchain.AddTransaction(transaction.Transaction{
		From:   transaction.Faucet,
		To:     "Alice",
		Amount: 100,
	})
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	_, _, err = blockchain.MinePending()
	if err != nil {
		t.Fatalf("failed to mine block: %v", err)
	}

	// Deliberately change the mined transaction.
	blockchain.Blocks[1].Transactions[0].Amount = 999

	result := blockchain.Validate()

	if result.Valid {
		t.Fatal("expected tampered blockchain to be invalid")
	}

	expectedReason := "stored Merkle root does not match block transactions"

	if result.Reason != expectedReason {
		t.Fatalf(
			"expected reason %q, got %q",
			expectedReason,
			result.Reason,
		)
	}
}

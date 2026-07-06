package block

import (
	"testing"

	"toy-blockchain/internal/transaction"
)

func TestGenesisBlockIsDeterministic(t *testing.T) {
	first := NewGenesisBlock()
	second := NewGenesisBlock()

	if first.Height != 0 {
		t.Fatalf("expected genesis height 0, got %d", first.Height)
	}

	if first.PrevHash != GenesisPrevHash {
		t.Fatalf("expected fixed genesis previous hash")
	}

	if first.Hash != second.Hash {
		t.Fatalf("expected deterministic genesis hash")
	}
}

func TestCalculateHashIsDeterministic(t *testing.T) {
	blk := Block{
		Height:    1,
		Timestamp: 123456789,
		Transactions: []transaction.Transaction{
			{From: transaction.Faucet, To: "Alice", Amount: 100},
		},
		PrevHash: "abc",
		Nonce:    42,
	}

	first := blk.CalculateHash()
	second := blk.CalculateHash()

	if first != second {
		t.Fatalf("expected same hash twice, got %s and %s", first, second)
	}
}

func TestMiningMeetsDifficulty(t *testing.T) {
	blk := NewBlock(1, []transaction.Transaction{{From: transaction.Faucet, To: "Alice", Amount: 100}}, NewGenesisBlock().Hash)
	result := blk.Mine(2)

	if !MeetsDifficulty(result.Hash, 2) {
		t.Fatalf("hash does not meet difficulty: %s", result.Hash)
	}

	if blk.CalculateHash() != result.Hash {
		t.Fatalf("nonce does not reproduce mined hash")
	}
}

package block

import (
	"testing"

	"toy-blockchain/internal/transaction"
)

func TestDifficultyIsIncludedInBlockHash(t *testing.T) {
	blk := Block{
		Height:    1,
		Timestamp: 123,
		Transactions: []transaction.Transaction{
			{
				From:   transaction.Faucet,
				To:     "Alice",
				Amount: 10,
			},
		},
		PrevHash:   "abc",
		Difficulty: 1,
		Nonce:      42,
	}

	blk.MerkleRoot = blk.CalculateMerkleRoot()

	firstHash := blk.CalculateHash()

	blk.Difficulty = 2
	secondHash := blk.CalculateHash()

	if firstHash == secondHash {
		t.Fatal(
			"expected changing difficulty to change the block hash",
		)
	}
}

func TestMiningStoresDifficultyInBlock(t *testing.T) {
	blk := NewBlock(
		1,
		[]transaction.Transaction{
			{
				From:   transaction.Faucet,
				To:     "Alice",
				Amount: 10,
			},
		},
		NewGenesisBlock().Hash,
	)

	blk.Mine(2)

	if blk.Difficulty != 2 {
		t.Fatalf(
			"expected stored difficulty 2, got %d",
			blk.Difficulty,
		)
	}
}

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
		t.Fatal("expected fixed genesis previous hash")
	}

	if first.MerkleRoot == "" {
		t.Fatal("expected genesis block to have a Merkle root")
	}

	if first.MerkleRoot != second.MerkleRoot {
		t.Fatal("expected deterministic genesis Merkle root")
	}

	if first.Hash != second.Hash {
		t.Fatal("expected deterministic genesis hash")
	}
}

func TestNewBlockCalculatesMerkleRoot(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
	}

	blk := NewBlock(1, transactions, NewGenesisBlock().Hash)

	if blk.MerkleRoot == "" {
		t.Fatal("expected new block to contain a Merkle root")
	}

	if blk.MerkleRoot != blk.CalculateMerkleRoot() {
		t.Fatal("stored Merkle root does not match transactions")
	}
}

func TestCalculateHashIsDeterministic(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
	}

	blk := Block{
		Height:       1,
		Timestamp:    123456789,
		Transactions: transactions,
		PrevHash:     "abc",
		Nonce:        42,
	}

	blk.MerkleRoot = blk.CalculateMerkleRoot()

	first := blk.CalculateHash()
	second := blk.CalculateHash()

	if first != second {
		t.Fatalf(
			"expected same hash twice, got %s and %s",
			first,
			second,
		)
	}
}

func TestChangingMerkleRootChangesBlockHash(t *testing.T) {
	transactions := []transaction.Transaction{
		{
			From:   transaction.Faucet,
			To:     "Alice",
			Amount: 100,
		},
	}

	blk := Block{
		Height:       1,
		Timestamp:    123456789,
		Transactions: transactions,
		PrevHash:     "abc",
		Nonce:        42,
	}

	blk.MerkleRoot = blk.CalculateMerkleRoot()
	originalHash := blk.CalculateHash()

	blk.MerkleRoot = "changed-root"
	modifiedHash := blk.CalculateHash()

	if originalHash == modifiedHash {
		t.Fatal("expected changed Merkle root to change block hash")
	}
}

func TestMiningMeetsDifficulty(t *testing.T) {
	blk := NewBlock(
		1,
		[]transaction.Transaction{
			{
				From:   transaction.Faucet,
				To:     "Alice",
				Amount: 100,
			},
		},
		NewGenesisBlock().Hash,
	)

	result := blk.Mine(2)

	if !MeetsDifficulty(result.Hash, 2) {
		t.Fatalf("hash does not meet difficulty: %s", result.Hash)
	}

	if blk.CalculateHash() != result.Hash {
		t.Fatal("nonce does not reproduce mined hash")
	}

	if blk.MerkleRoot != blk.CalculateMerkleRoot() {
		t.Fatal("mined block has an invalid Merkle root")
	}
}
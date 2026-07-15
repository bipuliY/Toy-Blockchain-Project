package chain

import (
	"fmt"
	"testing"

	"toy-blockchain/block"
	"toy-blockchain/internal/transaction"
)

func chainWithTimestamps(
	difficulty int,
	timestamps []int64,
) *Blockchain {
	bc := NewBlockchain(difficulty, 5)

	for i, timestamp := range timestamps {
		bc.Blocks = append(
			bc.Blocks,
			block.Block{
				Height:     i + 1,
				Timestamp:  timestamp,
				Difficulty: difficulty,
			},
		)
	}

	return bc
}

func TestDifficultyIncreasesWhenBlocksAreTooFast(
	t *testing.T,
) {
	bc := chainWithTimestamps(
		2,
		[]int64{100, 101, 102, 103, 104},
	)

	got := bc.CalculateNextDifficulty()

	if got != 3 {
		t.Fatalf(
			"expected difficulty 3, got %d",
			got,
		)
	}
}

func TestDifficultyDecreasesWhenBlocksAreTooSlow(
	t *testing.T,
) {
	bc := chainWithTimestamps(
		2,
		[]int64{100, 130, 160, 190, 220},
	)

	got := bc.CalculateNextDifficulty()

	if got != 1 {
		t.Fatalf(
			"expected difficulty 1, got %d",
			got,
		)
	}
}

func TestDifficultyStaysSameNearTarget(
	t *testing.T,
) {
	bc := chainWithTimestamps(
		2,
		[]int64{100, 110, 120, 130, 140},
	)

	got := bc.CalculateNextDifficulty()

	if got != 2 {
		t.Fatalf(
			"expected difficulty 2, got %d",
			got,
		)
	}
}

func TestRetargetedChainStillValidates(
	t *testing.T,
) {
	bc := NewBlockchain(1, 5)

	for i := 0; i < 6; i++ {
		tx := transaction.New(
			transaction.Faucet,
			fmt.Sprintf("User%d", i),
			10,
		)

		if err := bc.AddTransaction(tx); err != nil {
			t.Fatalf(
				"add transaction %d: %v",
				i,
				err,
			)
		}

		if _, _, err := bc.MinePending(); err != nil {
			t.Fatalf(
				"mine block %d: %v",
				i+1,
				err,
			)
		}
	}

	if bc.Blocks[5].Difficulty != 1 {
		t.Fatalf(
			"expected block 5 difficulty 1, got %d",
			bc.Blocks[5].Difficulty,
		)
	}

	if bc.Blocks[6].Difficulty != 2 {
		t.Fatalf(
			"expected block 6 difficulty 2, got %d",
			bc.Blocks[6].Difficulty,
		)
	}

	result := bc.Validate()

	if !result.Valid {
		t.Fatalf(
			"expected retargeted chain to validate, got %s",
			result.Reason,
		)
	}
}

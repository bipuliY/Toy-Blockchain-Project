package chain

import (
	"errors"
	"fmt"

	"toy-blockchain/block"
	"toy-blockchain/internal/transaction"
	"toy-blockchain/ledger"
)

const (
	DefaultDifficulty = 2
	DefaultBlockSize  = 5

	DefaultTargetBlockTimeSeconds       = 10
	DefaultRetargetInterval             = 5
	DefaultMinDifficulty                = 1
	DefaultMaxDifficulty                = 6
)

type Blockchain struct {
	Blocks                 []block.Block             `json:"blocks"`
	PendingTransactions    []transaction.Transaction `json:"pending_transactions"`
	Difficulty             int                       `json:"difficulty"`
	BlockSize              int                       `json:"block_size"`
	TargetBlockTimeSeconds int64                     `json:"target_block_time_seconds"`
	RetargetInterval       int                       `json:"retarget_interval"`
	MinDifficulty          int                       `json:"min_difficulty"`
	MaxDifficulty          int                       `json:"max_difficulty"`
}

type ValidationResult struct {
	Valid       bool   `json:"valid"`
	BlockHeight int    `json:"block_height"`
	Reason      string `json:"reason"`
}

func NewBlockchain(difficulty int, blockSize int) *Blockchain {
	if difficulty <= 0 {
		difficulty = DefaultDifficulty
	}

	if blockSize <= 0 {
		blockSize = DefaultBlockSize
	}

	if difficulty < DefaultMinDifficulty {
		difficulty = DefaultMinDifficulty
	}

	if difficulty > DefaultMaxDifficulty {
		difficulty = DefaultMaxDifficulty
	}

	return &Blockchain{
		Blocks:                 []block.Block{block.NewGenesisBlock()},
		PendingTransactions:    []transaction.Transaction{},
		Difficulty:             difficulty,
		BlockSize:              blockSize,
		TargetBlockTimeSeconds: DefaultTargetBlockTimeSeconds,
		RetargetInterval:       DefaultRetargetInterval,
		MinDifficulty:          DefaultMinDifficulty,
		MaxDifficulty:          DefaultMaxDifficulty,
	}
}

func (bc *Blockchain) AddTransaction(tx transaction.Transaction) error {
	balances := bc.BalancesIncludingPending()

	if err := ledger.ValidateTransaction(tx, balances); err != nil {
		return err
	}

	bc.PendingTransactions = append(bc.PendingTransactions, tx)
	return nil
}

func (bc *Blockchain) MinePending() (block.Block, block.MineResult, error) {
	if len(bc.Blocks) == 0 {
		return block.Block{}, block.MineResult{}, errors.New("blockchain has no genesis block")
	}

	if len(bc.PendingTransactions) == 0 {
		return block.Block{}, block.MineResult{}, errors.New("no pending transactions to mine")
	}

	txCount := len(bc.PendingTransactions)
	if txCount > bc.BlockSize {
		txCount = bc.BlockSize
	}

	txsToMine := append([]transaction.Transaction(nil), bc.PendingTransactions[:txCount]...)
	previousBlock := bc.Blocks[len(bc.Blocks)-1]

	// newBlock := block.NewBlock(len(bc.Blocks), txsToMine, previousBlock.Hash)
	// //mineResult := newBlock.Mine(bc.Difficulty)
	// mineResult := newBlock.MineConcurrent(bc.Difficulty, 0)

	// bc.Blocks = append(bc.Blocks, newBlock)
	// bc.PendingTransactions = bc.PendingTransactions[txCount:]
	newBlock := block.NewBlock(
		len(bc.Blocks),
		txsToMine,
		previousBlock.Hash,
	)

	mineResult := newBlock.MineConcurrent(
		bc.Difficulty,
		0,
	)

	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingTransactions = bc.PendingTransactions[txCount:]

	// Calculate the difficulty that should be used by the next block.
	bc.Difficulty = bc.CalculateNextDifficulty()

	return newBlock, mineResult, nil
}

func (bc *Blockchain) Balances() map[string]int {
	return ledger.Balances(bc.Blocks)
}

func (bc *Blockchain) BalancesIncludingPending() map[string]int {
	balances := bc.Balances()

	for _, tx := range bc.PendingTransactions {
		ledger.ApplyTransaction(balances, tx)
	}

	return balances
}
func clampDifficulty(value int, minimum int, maximum int) int {
	if minimum <= 0 {
		minimum = 1
	}

	if maximum < minimum {
		maximum = minimum
	}

	if value < minimum {
		return minimum
	}

	if value > maximum {
		return maximum
	}

	return value
}

func calculateRetargetedDifficulty(
	blocks []block.Block,
	currentDifficulty int,
	targetBlockTimeSeconds int64,
	retargetInterval int,
	minDifficulty int,
	maxDifficulty int,
) int {
	currentDifficulty = clampDifficulty(
		currentDifficulty,
		minDifficulty,
		maxDifficulty,
	)

	minedBlockCount := len(blocks) - 1

	if targetBlockTimeSeconds <= 0 {
		return currentDifficulty
	}

	if retargetInterval < 2 {
		return currentDifficulty
	}

	if minedBlockCount < retargetInterval {
		return currentDifficulty
	}

	if minedBlockCount%retargetInterval != 0 {
		return currentDifficulty
	}

	endIndex := len(blocks) - 1
	startIndex := endIndex - (retargetInterval - 1)

	actualDuration :=
		blocks[endIndex].Timestamp -
			blocks[startIndex].Timestamp

	expectedDuration :=
		int64(retargetInterval-1) *
			targetBlockTimeSeconds

	// Blocks were produced in less than half the expected time.
	if actualDuration*2 < expectedDuration {
		return clampDifficulty(
			currentDifficulty+1,
			minDifficulty,
			maxDifficulty,
		)
	}

	// Blocks took more than twice the expected time.
	if actualDuration > expectedDuration*2 {
		return clampDifficulty(
			currentDifficulty-1,
			minDifficulty,
			maxDifficulty,
		)
	}

	return currentDifficulty
}

func (bc *Blockchain) CalculateNextDifficulty() int {
	return calculateRetargetedDifficulty(
		bc.Blocks,
		bc.Difficulty,
		bc.TargetBlockTimeSeconds,
		bc.RetargetInterval,
		bc.MinDifficulty,
		bc.MaxDifficulty,
	)
}

func (bc *Blockchain) Validate() ValidationResult {
	if len(bc.Blocks) == 0 {
		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason:      "chain has no blocks",
		}
	}

	if bc.TargetBlockTimeSeconds <= 0 {
		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason:      "target block time must be positive",
		}
	}

	if bc.RetargetInterval < 2 {
		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason:      "retarget interval must be at least 2",
		}
	}

	if bc.MinDifficulty <= 0 ||
		bc.MaxDifficulty < bc.MinDifficulty {

		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason:      "difficulty limits are invalid",
		}
	}

	balances := map[string]int{}

	expectedDifficulty := bc.Difficulty

	if len(bc.Blocks) > 1 {
		expectedDifficulty = bc.Blocks[1].Difficulty
	}

	if expectedDifficulty < bc.MinDifficulty ||
		expectedDifficulty > bc.MaxDifficulty {

		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason:      "initial block difficulty is outside allowed limits",
		}
	}

	for i, current := range bc.Blocks {
		if current.Height != i {
			return ValidationResult{
				Valid:       false,
				BlockHeight: current.Height,
				Reason: fmt.Sprintf(
					"invalid height: expected %d but found %d",
					i,
					current.Height,
				),
			}
		}

		calculatedMerkleRoot := current.CalculateMerkleRoot()

		if current.MerkleRoot != calculatedMerkleRoot {
			return ValidationResult{
				Valid:       false,
				BlockHeight: current.Height,
				Reason:      "stored Merkle root does not match block transactions",
			}
		}

		calculatedHash := current.CalculateHash()

		if current.Hash != calculatedHash {
			return ValidationResult{
				Valid:       false,
				BlockHeight: current.Height,
				Reason:      "stored hash does not match recalculated hash",
			}
		}

		if i == 0 {
			if current.PrevHash != block.GenesisPrevHash {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "genesis previous hash is invalid",
				}
			}

			if current.Difficulty != 0 {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "genesis difficulty must be zero",
				}
			}
		} else {
			previous := bc.Blocks[i-1]

			if current.PrevHash != previous.Hash {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "previous hash link is broken",
				}
			}

			if current.Timestamp < previous.Timestamp {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "timestamp is older than previous block",
				}
			}

			if current.Difficulty != expectedDifficulty {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason: fmt.Sprintf(
						"unexpected difficulty: expected %d but found %d",
						expectedDifficulty,
						current.Difficulty,
					),
				}
			}

			if !block.MeetsDifficulty(
				current.Hash,
				current.Difficulty,
			) {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "block hash does not meet its proof-of-work difficulty",
				}
			}
		}

		for _, tx := range current.Transactions {
			if err := ledger.ValidateTransaction(tx, balances); err != nil {
				return ValidationResult{
					Valid:       false,
					BlockHeight: current.Height,
					Reason:      "invalid transaction: " + err.Error(),
				}
			}

			ledger.ApplyTransaction(balances, tx)
		}

		// Determine the expected difficulty for the next block.
		if i > 0 && i%bc.RetargetInterval == 0 {
			expectedDifficulty = calculateRetargetedDifficulty(
				bc.Blocks[:i+1],
				expectedDifficulty,
				bc.TargetBlockTimeSeconds,
				bc.RetargetInterval,
				bc.MinDifficulty,
				bc.MaxDifficulty,
			)
		}
	}

	if bc.Difficulty != expectedDifficulty {
		return ValidationResult{
			Valid:       false,
			BlockHeight: -1,
			Reason: fmt.Sprintf(
				"stored next difficulty is incorrect: expected %d but found %d",
				expectedDifficulty,
				bc.Difficulty,
			),
		}
	}

	return ValidationResult{
		Valid:       true,
		BlockHeight: -1,
		Reason:      "chain is valid",
	}
}

// func (bc *Blockchain) Validate() ValidationResult {
// 	if len(bc.Blocks) == 0 {
// 		return ValidationResult{Valid: false, BlockHeight: -1, Reason: "chain has no blocks"}
// 	}

// 	balances := map[string]int{}

// 	for i, current := range bc.Blocks {
// 		if current.Height != i {
// 			return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: fmt.Sprintf("invalid height: expected %d but found %d", i, current.Height)}
// 		}

// 		// calculatedHash := current.CalculateHash()
// 		// if current.Hash != calculatedHash {
// 		// 	return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "stored hash does not match recalculated hash"}
// 		// }
// 		calculatedMerkleRoot := current.CalculateMerkleRoot()

// 		if current.MerkleRoot != calculatedMerkleRoot {
// 			return ValidationResult{
// 				Valid:       false,
// 				BlockHeight: current.Height,
// 				Reason:      "stored Merkle root does not match block transactions",
// 			}
// 		}

// 		calculatedHash := current.CalculateHash()

// 		if current.Hash != calculatedHash {
// 			return ValidationResult{
// 				Valid:       false,
// 				BlockHeight: current.Height,
// 				Reason:      "stored hash does not match recalculated hash",
// 			}
// 		}

// 		if i == 0 {
// 			if current.PrevHash != block.GenesisPrevHash {
// 				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "genesis previous hash is invalid"}
// 			}
// 		} else {
// 			previous := bc.Blocks[i-1]

// 			if current.PrevHash != previous.Hash {
// 				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "previous hash link is broken"}
// 			}

// 			if current.Timestamp < previous.Timestamp {
// 				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "timestamp is older than previous block"}
// 			}

// 			if !block.MeetsDifficulty(current.Hash, bc.Difficulty) {
// 				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "block hash does not meet proof-of-work difficulty"}
// 			}
// 		}

// 		for _, tx := range current.Transactions {
// 			if err := ledger.ValidateTransaction(tx, balances); err != nil {
// 				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "invalid transaction: " + err.Error()}
// 			}

// 			ledger.ApplyTransaction(balances, tx)
// 		}
// 	}

// 	return ValidationResult{Valid: true, BlockHeight: -1, Reason: "chain is valid"}
// }

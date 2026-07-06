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
)

type Blockchain struct {
	Blocks              []block.Block             `json:"blocks"`
	PendingTransactions []transaction.Transaction `json:"pending_transactions"`
	Difficulty          int                       `json:"difficulty"`
	BlockSize           int                       `json:"block_size"`
}

type ValidationResult struct {
	Valid       bool   `json:"valid"`
	BlockHeight int    `json:"block_height"`
	Reason      string `json:"reason"`
}

func NewBlockchain(difficulty int, blockSize int) *Blockchain {
	if difficulty < 0 {
		difficulty = DefaultDifficulty
	}

	if blockSize <= 0 {
		blockSize = DefaultBlockSize
	}

	return &Blockchain{
		Blocks:              []block.Block{block.NewGenesisBlock()},
		PendingTransactions: []transaction.Transaction{},
		Difficulty:          difficulty,
		BlockSize:           blockSize,
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

	newBlock := block.NewBlock(len(bc.Blocks), txsToMine, previousBlock.Hash)
	mineResult := newBlock.Mine(bc.Difficulty)

	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingTransactions = bc.PendingTransactions[txCount:]

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

func (bc *Blockchain) Validate() ValidationResult {
	if len(bc.Blocks) == 0 {
		return ValidationResult{Valid: false, BlockHeight: -1, Reason: "chain has no blocks"}
	}

	balances := map[string]int{}

	for i, current := range bc.Blocks {
		if current.Height != i {
			return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: fmt.Sprintf("invalid height: expected %d but found %d", i, current.Height)}
		}

		calculatedHash := current.CalculateHash()
		if current.Hash != calculatedHash {
			return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "stored hash does not match recalculated hash"}
		}

		if i == 0 {
			if current.PrevHash != block.GenesisPrevHash {
				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "genesis previous hash is invalid"}
			}
		} else {
			previous := bc.Blocks[i-1]

			if current.PrevHash != previous.Hash {
				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "previous hash link is broken"}
			}

			if current.Timestamp < previous.Timestamp {
				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "timestamp is older than previous block"}
			}

			if !block.MeetsDifficulty(current.Hash, bc.Difficulty) {
				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "block hash does not meet proof-of-work difficulty"}
			}
		}

		for _, tx := range current.Transactions {
			if err := ledger.ValidateTransaction(tx, balances); err != nil {
				return ValidationResult{Valid: false, BlockHeight: current.Height, Reason: "invalid transaction: " + err.Error()}
			}

			ledger.ApplyTransaction(balances, tx)
		}
	}

	return ValidationResult{Valid: true, BlockHeight: -1, Reason: "chain is valid"}
}

package block

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"toy-blockchain/internal/transaction"
	"toy-blockchain/merkle"
)

const GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"

type Block struct {
	Height       int                       `json:"height"`
	Timestamp    int64                     `json:"timestamp"`
	Transactions []transaction.Transaction `json:"transactions"`
	MerkleRoot   string                    `json:"merkle_root"`
	PrevHash     string                    `json:"prev_hash"`
	Difficulty   int                       `json:"difficulty"`
	Nonce        int                       `json:"nonce"`
	Hash         string                    `json:"hash"`
}

type MineResult struct {
	Nonce          int    `json:"nonce"`
	Hash           string `json:"hash"`
	DurationMillis int64  `json:"duration_millis"`
	HashesTried    int64  `json:"hashes_tried"`
}
type concurrentMineResult struct {
	Nonce int
	Hash  string
}

// hashInput contains only the values used to calculate the block hash.
//
// The raw transaction list is no longer included.
// MerkleRoot represents all transactions in the block.
type hashInput struct {
	Height     int    `json:"height"`
	Timestamp  int64  `json:"timestamp"`
	MerkleRoot string `json:"merkle_root"`
	PrevHash   string `json:"prev_hash"`
	Difficulty int    `json:"difficulty"`
	Nonce      int    `json:"nonce"`
}

func NewGenesisBlock() Block {
	transactions := []transaction.Transaction{}

	genesis := Block{
		Height:       0,
		Timestamp:    0,
		Transactions: transactions,
		MerkleRoot:   merkle.CalculateRoot(transactions),
		PrevHash:     GenesisPrevHash,
		Difficulty:   0,
		Nonce:        0,
	}

	genesis.Hash = genesis.CalculateHash()
	return genesis
}

func NewBlock(
	height int,
	transactions []transaction.Transaction,
	prevHash string,
) Block {
	// Copy transactions so that the block owns its own transaction slice.
	transactionCopy := append([]transaction.Transaction(nil), transactions...)

	return Block{
		Height:       height,
		Timestamp:    time.Now().Unix(),
		Transactions: transactionCopy,
		MerkleRoot:   merkle.CalculateRoot(transactionCopy),
		PrevHash:     prevHash,
		Nonce:        0,
	}
}

// func (b Block) CalculateHash() string {
// 	input := hashInput{
// 		Height:     b.Height,
// 		Timestamp:  b.Timestamp,
// 		MerkleRoot: b.MerkleRoot,
// 		PrevHash:   b.PrevHash,
// 		Nonce:      b.Nonce,
// 	}

// 	bytes, err := json.Marshal(input)
// 	if err != nil {
// 		panic(err)
// 	}

// 	sum := sha256.Sum256(bytes)
// 	return hex.EncodeToString(sum[:])
// }

func (b Block) CalculateHash() string {
	return b.calculateHashForNonce(b.Nonce)
}

func (b Block) calculateHashForNonce(nonce int) string {
	input := hashInput{
		Height:     b.Height,
		Timestamp:  b.Timestamp,
		MerkleRoot: b.MerkleRoot,
		PrevHash:   b.PrevHash,
		Difficulty: b.Difficulty,
		Nonce:      nonce,
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

// CalculateMerkleRoot recalculates the root from the block's transactions.
func (b Block) CalculateMerkleRoot() string {
	return merkle.CalculateRoot(b.Transactions)
}
func (b *Block) MineConcurrent(difficulty int, workerCount int) MineResult {
	start := time.Now()

	// Make sure the block contains the correct transaction summary.
	b.MerkleRoot = b.CalculateMerkleRoot()

	if difficulty < 1 {
		difficulty = 1
	}

	b.Difficulty = difficulty

	// Use the available CPU count when no valid worker count is provided.
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChannel := make(chan concurrentMineResult, 1)

	var waitGroup sync.WaitGroup
	var hashesTried int64

	// Copy the block so workers only read stable block information.
	blockCopy := *b

	for workerID := 0; workerID < workerCount; workerID++ {
		waitGroup.Add(1)

		go func(id int) {
			defer waitGroup.Done()

			// Each worker starts from a different nonce.
			nonce := id

			for {
				// Stop when another worker has found a valid result.
				select {
				case <-ctx.Done():
					return
				default:
				}

				atomic.AddInt64(&hashesTried, 1)

				hash := blockCopy.calculateHashForNonce(nonce)

				if MeetsDifficulty(hash, difficulty) {
					// Only the first successful result is accepted.
					select {
					case resultChannel <- concurrentMineResult{
						Nonce: nonce,
						Hash:  hash,
					}:
						cancel()
					case <-ctx.Done():
					}

					return
				}

				// Move to this worker's next nonce.
				nonce += workerCount
			}
		}(workerID)
	}

	// Wait for the first valid result.
	winner := <-resultChannel

	// Make sure every worker stops before returning.
	cancel()
	waitGroup.Wait()

	b.Nonce = winner.Nonce
	b.Hash = winner.Hash

	return MineResult{
		Nonce:          winner.Nonce,
		Hash:           winner.Hash,
		DurationMillis: time.Since(start).Milliseconds(),
		HashesTried:    atomic.LoadInt64(&hashesTried),
	}
}
func (b *Block) Mine(difficulty int) MineResult {
	start := time.Now()
	var tries int64

	// Always calculate the root before mining.
	b.MerkleRoot = b.CalculateMerkleRoot()

	if difficulty < 1 {
		difficulty = 1
	}

	b.Difficulty = difficulty

	for {
		tries++

		hash := b.CalculateHash()

		if MeetsDifficulty(hash, difficulty) {
			b.Hash = hash

			return MineResult{
				Nonce:          b.Nonce,
				Hash:           hash,
				DurationMillis: time.Since(start).Milliseconds(),
				HashesTried:    tries,
			}
		}

		b.Nonce++
	}
}

func MeetsDifficulty(hash string, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}

	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

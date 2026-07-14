package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
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
	Nonce        int                       `json:"nonce"`
	Hash         string                    `json:"hash"`
}

type MineResult struct {
	Nonce          int    `json:"nonce"`
	Hash           string `json:"hash"`
	DurationMillis int64  `json:"duration_millis"`
	HashesTried    int64  `json:"hashes_tried"`
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

func (b Block) CalculateHash() string {
	input := hashInput{
		Height:     b.Height,
		Timestamp:  b.Timestamp,
		MerkleRoot: b.MerkleRoot,
		PrevHash:   b.PrevHash,
		Nonce:      b.Nonce,
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

func (b *Block) Mine(difficulty int) MineResult {
	start := time.Now()
	var tries int64

	// Always calculate the root before mining.
	b.MerkleRoot = b.CalculateMerkleRoot()

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
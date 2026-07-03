package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"toy-blockchain/internal/transaction"
)

const GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"

type Block struct {
	Height       int                       `json:"height"`
	Timestamp    int64                     `json:"timestamp"`
	Transactions []transaction.Transaction `json:"transactions"`
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

type hashInput struct {
	Height       int                       `json:"height"`
	Timestamp    int64                     `json:"timestamp"`
	Transactions []transaction.Transaction `json:"transactions"`
	PrevHash     string                    `json:"prev_hash"`
	Nonce        int                       `json:"nonce"`
}

func NewGenesisBlock() Block {
	genesis := Block{
		Height:       0,
		Timestamp:    0,
		Transactions: []transaction.Transaction{},
		PrevHash:     GenesisPrevHash,
		Nonce:        0,
	}

	genesis.Hash = genesis.CalculateHash()
	return genesis
}

func NewBlock(height int, transactions []transaction.Transaction, prevHash string) Block {
	return Block{
		Height:       height,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
	}
}

func (b Block) CalculateHash() string {
	input := hashInput{
		Height:       b.Height,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
		Nonce:        b.Nonce,
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

func (b *Block) Mine(difficulty int) MineResult {
	start := time.Now()
	var tries int64

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
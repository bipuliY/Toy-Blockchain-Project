package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"toy-blockchain/internal/transaction"
)

// CalculateRoot calculates one Merkle root representing all transactions.
func CalculateRoot(transactions []transaction.Transaction) string {
	// A block may contain no transactions, such as the genesis block.
	// SHA-256 of empty bytes gives us a fixed deterministic root.
	if len(transactions) == 0 {
		emptyHash := sha256.Sum256([]byte{})
		return hex.EncodeToString(emptyHash[:])
	}

	// First create one hash for each transaction.
	currentLevel := make([][]byte, 0, len(transactions))

	for _, tx := range transactions {
		txHash := hashTransaction(tx)
		currentLevel = append(currentLevel, txHash)
	}

	// Continue combining hashes until only one hash remains.
	for len(currentLevel) > 1 {
		// If the number of hashes is odd, duplicate the last hash.
		if len(currentLevel)%2 != 0 {
			lastHashCopy := append([]byte(nil), currentLevel[len(currentLevel)-1]...)
			currentLevel = append(currentLevel, lastHashCopy)
		}

		nextLevel := make([][]byte, 0, len(currentLevel)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			combined := make([]byte, 0, len(currentLevel[i])+len(currentLevel[i+1]))
			combined = append(combined, currentLevel[i]...)
			combined = append(combined, currentLevel[i+1]...)

			parentHash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, parentHash[:])
		}

		currentLevel = nextLevel
	}

	return hex.EncodeToString(currentLevel[0])
}

// hashTransaction creates a deterministic SHA-256 hash for one transaction.
func hashTransaction(tx transaction.Transaction) []byte {
	transactionBytes, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(transactionBytes)
	return hash[:]
}
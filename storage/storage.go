package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"toy-blockchain/chain"
)

func Load(path string) (*chain.Blockchain, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bc chain.Blockchain
	if err := json.Unmarshal(bytes, &bc); err != nil {
		return nil, err
	}

	if bc.BlockSize <= 0 {
		bc.BlockSize = chain.DefaultBlockSize
	}

	if bc.Difficulty < 0 {
		bc.Difficulty = chain.DefaultDifficulty
	}

	return &bc, nil
}

func Save(path string, bc *chain.Blockchain) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

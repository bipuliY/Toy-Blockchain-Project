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

	// If difficulty is missing or non-positive in the JSON, use the default.
	if bc.Difficulty <= 0 {
		bc.Difficulty = chain.DefaultDifficulty
	}

	if bc.TargetBlockTimeSeconds <= 0 {
		bc.TargetBlockTimeSeconds =
			chain.DefaultTargetBlockTimeSeconds
	}

	if bc.RetargetInterval < 2 {
		bc.RetargetInterval =
			chain.DefaultRetargetInterval
	}

	if bc.MinDifficulty <= 0 {
		bc.MinDifficulty =
			chain.DefaultMinDifficulty
	}

	if bc.MaxDifficulty < bc.MinDifficulty {
		bc.MaxDifficulty =
			chain.DefaultMaxDifficulty
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

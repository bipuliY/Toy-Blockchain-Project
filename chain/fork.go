package chain

import (
	"fmt"

	"toy-blockchain/block"
	"toy-blockchain/internal/transaction"
)

// CandidateEvaluation stores the result of checking one competing chain.
type CandidateEvaluation struct {
	Index    int    `json:"index"`
	Length   int    `json:"length"`
	Valid    bool   `json:"valid"`
	Selected bool   `json:"selected"`
	Reason   string `json:"reason"`
}

// ForkResolutionResult stores the final fork-resolution decision.
type ForkResolutionResult struct {
	Replaced          bool                  `json:"replaced"`
	PreviousLength    int                   `json:"previous_length"`
	SelectedLength    int                   `json:"selected_length"`
	SelectedCandidate int                   `json:"selected_candidate"`
	RetainedPending   int                   `json:"retained_pending"`
	DroppedPending    int                   `json:"dropped_pending"`
	Reason            string                `json:"reason"`
	Candidates        []CandidateEvaluation `json:"candidates"`
}

// ResolveFork checks competing chains and selects the longest valid chain.
func (bc *Blockchain) ResolveFork(
	candidates []*Blockchain,
) ForkResolutionResult {
	result := ForkResolutionResult{
		SelectedCandidate: -1,
		Candidates:        make([]CandidateEvaluation, 0, len(candidates)),
	}

	if bc == nil {
		result.Reason = "local blockchain is nil"
		return result
	}

	result.PreviousLength = len(bc.Blocks)
	result.SelectedLength = len(bc.Blocks)

	localValidation := bc.Validate()

	bestLength := 0
	if localValidation.Valid {
		bestLength = len(bc.Blocks)
	}

	selectedCandidate := -1

	for index, candidate := range candidates {
		evaluation := CandidateEvaluation{
			Index: index,
		}

		if candidate == nil {
			evaluation.Reason = "candidate is nil"
			result.Candidates = append(result.Candidates, evaluation)
			continue
		}

		evaluation.Length = len(candidate.Blocks)

		// Ensure that both chains belong to the same blockchain network.
		if err := bc.ensureForkCompatible(candidate); err != nil {
			evaluation.Reason =
				"incompatible candidate: " + err.Error()

			result.Candidates = append(
				result.Candidates,
				evaluation,
			)

			continue
		}

		// Reuse the existing full-chain validation.
		validation := candidate.Validate()

		if !validation.Valid {
			evaluation.Reason = fmt.Sprintf(
				"invalid candidate at block %d: %s",
				validation.BlockHeight,
				validation.Reason,
			)

			result.Candidates = append(
				result.Candidates,
				evaluation,
			)

			continue
		}

		evaluation.Valid = true
		evaluation.Reason = "candidate is valid"

		result.Candidates = append(
			result.Candidates,
			evaluation,
		)

		// Only a strictly longer chain can replace the current chain.
		// Equal-length chains do not replace the local chain.
		if len(candidate.Blocks) > bestLength {
			bestLength = len(candidate.Blocks)
			selectedCandidate = index
		}
	}

	// No better candidate was found.
	if selectedCandidate < 0 {
		if localValidation.Valid {
			result.Reason =
				"local chain remains the longest valid chain"
		} else {
			result.Reason =
				"local chain is invalid and no valid candidate was found"
		}

		return result
	}

	winner := candidates[selectedCandidate]

	// Keep a copy of the local pending transactions.
	localPending := append(
		[]transaction.Transaction(nil),
		bc.PendingTransactions...,
	)

	// Copy the winner into the local blockchain.
	adoptedChain := cloneBlockchain(winner)

	// A candidate node's pending transactions must not be copied.
	adoptedChain.PendingTransactions = nil

	*bc = *adoptedChain

	// Recheck local pending transactions against the selected chain.
	retained, dropped := restoreLocalPending(
		bc,
		localPending,
	)

	result.Candidates[selectedCandidate].Selected = true
	result.Replaced = true
	result.SelectedCandidate = selectedCandidate
	result.SelectedLength = len(bc.Blocks)
	result.RetainedPending = retained
	result.DroppedPending = dropped
	result.Reason = fmt.Sprintf(
		"candidate %d selected as the longest valid chain",
		selectedCandidate,
	)

	return result
}

// ensureForkCompatible checks whether two chains use the same network rules.
func (bc *Blockchain) ensureForkCompatible(
	candidate *Blockchain,
) error {
	if len(bc.Blocks) == 0 {
		return fmt.Errorf(
			"local chain has no genesis block",
		)
	}

	if len(candidate.Blocks) == 0 {
		return fmt.Errorf(
			"candidate chain has no genesis block",
		)
	}

	if bc.Blocks[0].Hash != candidate.Blocks[0].Hash {
		return fmt.Errorf(
			"genesis block does not match",
		)
	}

	if bc.BlockSize != candidate.BlockSize {
		return fmt.Errorf(
			"block size does not match",
		)
	}

	if bc.TargetBlockTimeSeconds !=
		candidate.TargetBlockTimeSeconds {
		return fmt.Errorf(
			"target block time does not match",
		)
	}

	if bc.RetargetInterval != candidate.RetargetInterval {
		return fmt.Errorf(
			"difficulty retarget interval does not match",
		)
	}

	if bc.MinDifficulty != candidate.MinDifficulty ||
		bc.MaxDifficulty != candidate.MaxDifficulty {
		return fmt.Errorf(
			"difficulty limits do not match",
		)
	}

	if initialDifficulty(bc) != initialDifficulty(candidate) {
		return fmt.Errorf(
			"initial difficulty does not match",
		)
	}

	return nil
}

// initialDifficulty returns the difficulty used by the first mined block.
func initialDifficulty(bc *Blockchain) int {
	if len(bc.Blocks) > 1 {
		return bc.Blocks[1].Difficulty
	}

	return bc.Difficulty
}

// cloneBlockchain creates an independent copy of a blockchain.
func cloneBlockchain(source *Blockchain) *Blockchain {
	blocks := make([]block.Block, len(source.Blocks))

	for index, sourceBlock := range source.Blocks {
		blocks[index] = sourceBlock

		blocks[index].Transactions = append(
			[]transaction.Transaction(nil),
			sourceBlock.Transactions...,
		)
	}

	return &Blockchain{
		Blocks:                 blocks,
		PendingTransactions:    nil,
		Difficulty:             source.Difficulty,
		BlockSize:              source.BlockSize,
		TargetBlockTimeSeconds: source.TargetBlockTimeSeconds,
		RetargetInterval:       source.RetargetInterval,
		MinDifficulty:          source.MinDifficulty,
		MaxDifficulty:          source.MaxDifficulty,
	}
}

// restoreLocalPending checks whether old local pending transactions
// are still valid after changing to the winning chain.
func restoreLocalPending(
	bc *Blockchain,
	pending []transaction.Transaction,
) (retained int, dropped int) {
	confirmedTransactions :=
		make(map[transaction.Transaction]int)

	for _, currentBlock := range bc.Blocks {
		for _, tx := range currentBlock.Transactions {
			confirmedTransactions[tx]++
		}
	}

	bc.PendingTransactions = nil

	for _, tx := range pending {
		// Do not add a transaction that is already confirmed.
		if confirmedTransactions[tx] > 0 {
			confirmedTransactions[tx]--
			dropped++
			continue
		}

		// AddTransaction checks balances and transaction validity again.
		if err := bc.AddTransaction(tx); err != nil {
			dropped++
			continue
		}

		retained++
	}

	return retained, dropped
}
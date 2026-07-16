package chain

import (
	"testing"

	"toy-blockchain/internal/transaction"
)

func mineFaucetBlock(
	t *testing.T,
	bc *Blockchain,
	to string,
	amount int,
) {
	t.Helper()

	tx := transaction.New(
		transaction.Faucet,
		to,
		amount,
	)

	if err := bc.AddTransaction(tx); err != nil {
		t.Fatalf(
			"failed to add faucet transaction: %v",
			err,
		)
	}

	if _, _, err := bc.MinePending(); err != nil {
		t.Fatalf(
			"failed to mine block: %v",
			err,
		)
	}
}

func TestResolveForkAdoptsLongerValidChain(
	t *testing.T,
) {
	local := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		local,
		"Alice",
		100,
	)

	candidate := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		candidate,
		"Bob",
		50,
	)

	mineFaucetBlock(
		t,
		candidate,
		"Carol",
		25,
	)

	result := local.ResolveFork(
		[]*Blockchain{candidate},
	)

	if !result.Replaced {
		t.Fatalf(
			"expected local chain to be replaced: %s",
			result.Reason,
		)
	}

	if len(local.Blocks) != len(candidate.Blocks) {
		t.Fatalf(
			"expected %d blocks, got %d",
			len(candidate.Blocks),
			len(local.Blocks),
		)
	}

	localTip :=
		local.Blocks[len(local.Blocks)-1].Hash

	candidateTip :=
		candidate.Blocks[len(candidate.Blocks)-1].Hash

	if localTip != candidateTip {
		t.Fatal(
			"candidate tip was not adopted",
		)
	}
}

func TestResolveForkRejectsLongerInvalidChain(
	t *testing.T,
) {
	local := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		local,
		"Alice",
		100,
	)

	candidate := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		candidate,
		"Bob",
		50,
	)

	mineFaucetBlock(
		t,
		candidate,
		"Carol",
		25,
	)

	// Deliberately tamper with the candidate.
	candidate.Blocks[1].Transactions[0].Amount = 999

	originalTip :=
		local.Blocks[len(local.Blocks)-1].Hash

	result := local.ResolveFork(
		[]*Blockchain{candidate},
	)

	if result.Replaced {
		t.Fatal(
			"invalid candidate should be rejected",
		)
	}

	currentTip :=
		local.Blocks[len(local.Blocks)-1].Hash

	if currentTip != originalTip {
		t.Fatal(
			"local chain changed after invalid candidate",
		)
	}

	if result.Candidates[0].Valid {
		t.Fatal(
			"candidate should be reported as invalid",
		)
	}
}

func TestResolveForkKeepsLocalChainOnEqualLength(
	t *testing.T,
) {
	local := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		local,
		"Alice",
		100,
	)

	candidate := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		candidate,
		"Bob",
		50,
	)

	originalTip :=
		local.Blocks[len(local.Blocks)-1].Hash

	result := local.ResolveFork(
		[]*Blockchain{candidate},
	)

	if result.Replaced {
		t.Fatal(
			"equal-length candidate should not replace local chain",
		)
	}

	currentTip :=
		local.Blocks[len(local.Blocks)-1].Hash

	if currentTip != originalTip {
		t.Fatal(
			"local chain changed during equal-length tie",
		)
	}
}

func TestResolveForkRejectsIncompatibleChain(
	t *testing.T,
) {
	local := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		local,
		"Alice",
		100,
	)

	// Different block size.
	candidate := NewBlockchain(1, 10)

	mineFaucetBlock(
		t,
		candidate,
		"Bob",
		50,
	)

	mineFaucetBlock(
		t,
		candidate,
		"Carol",
		25,
	)

	result := local.ResolveFork(
		[]*Blockchain{candidate},
	)

	if result.Replaced {
		t.Fatal(
			"incompatible candidate should be rejected",
		)
	}

	if result.Candidates[0].Valid {
		t.Fatal(
			"incompatible candidate should not be valid",
		)
	}
}

func TestResolveForkDropsInvalidPendingTransaction(
	t *testing.T,
) {
	local := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		local,
		"Alice",
		100,
	)

	pendingTransaction :=
		transaction.New("Alice", "Bob", 70)

	if err := local.AddTransaction(
		pendingTransaction,
	); err != nil {
		t.Fatalf(
			"failed to add pending transaction: %v",
			err,
		)
	}

	candidate := NewBlockchain(1, 5)

	mineFaucetBlock(
		t,
		candidate,
		"Alice",
		100,
	)

	candidateSpend :=
		transaction.New("Alice", "Carol", 40)

	if err := candidate.AddTransaction(
		candidateSpend,
	); err != nil {
		t.Fatalf(
			"failed to add candidate spend: %v",
			err,
		)
	}

	if _, _, err := candidate.MinePending(); err != nil {
		t.Fatalf(
			"failed to mine candidate spend: %v",
			err,
		)
	}

	result := local.ResolveFork(
		[]*Blockchain{candidate},
	)

	if !result.Replaced {
		t.Fatalf(
			"expected longer chain to be selected: %s",
			result.Reason,
		)
	}

	if len(local.PendingTransactions) != 0 {
		t.Fatalf(
			"expected pending transaction to be dropped, got %d",
			len(local.PendingTransactions),
		)
	}

	if result.DroppedPending != 1 {
		t.Fatalf(
			"expected one dropped transaction, got %d",
			result.DroppedPending,
		)
	}
}
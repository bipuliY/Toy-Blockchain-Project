package ledger

import (
	"fmt"

	"toy-blockchain/block"
	"toy-blockchain/internal/transaction"
)

func Balances(blocks []block.Block) map[string]int {
	balances := map[string]int{}

	for _, blk := range blocks {
		for _, tx := range blk.Transactions {
			ApplyTransaction(balances, tx)
		}
	}

	return balances
}

func ApplyTransaction(balances map[string]int, tx transaction.Transaction) {
	if tx.IsFaucet() {
		balances[tx.To] += tx.Amount
		return
	}

	balances[tx.From] -= tx.Amount
	balances[tx.To] += tx.Amount
}

func ValidateTransaction(tx transaction.Transaction, balances map[string]int) error {
	if err := tx.BasicValidate(); err != nil {
		return err
	}

	if tx.IsFaucet() {
		return nil
	}

	if balances[tx.From] < tx.Amount {
		return fmt.Errorf("insufficient balance: %s has %d but tried to send %d", tx.From, balances[tx.From], tx.Amount)
	}

	return nil
}
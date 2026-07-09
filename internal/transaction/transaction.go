package transaction

import (
	"errors"
	"strings"
)

const Faucet = "FAUCET"

type Transaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func New(from, to string, amount int) Transaction {
	return Transaction{
		From:   strings.TrimSpace(from),
		To:     strings.TrimSpace(to),
		Amount: amount,
	}
}

func (tx Transaction) IsFaucet() bool {
	return strings.EqualFold(tx.From, Faucet)
}

func (tx Transaction) BasicValidate() error {
	if strings.TrimSpace(tx.From) == "" {
		return errors.New("error:sender is required")
	}

	if strings.TrimSpace(tx.To) == "" {
		return errors.New("error:recipient is required")
	}

	if tx.Amount <= 0 {
		return errors.New("error:amount must be greater than zero")
	}

	if strings.EqualFold(strings.TrimSpace(tx.From), strings.TrimSpace(tx.To)) {
		return errors.New("error:sender and recipient cannot be the same")
	}

	return nil
}

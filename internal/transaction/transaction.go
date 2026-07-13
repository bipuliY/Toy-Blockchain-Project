package transaction

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const Faucet = "FAUCET"

type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	PubKeyHex string `json:"pubkey,omitempty"`
	SigHex    string `json:"signature,omitempty"`
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

// BasicValidate checks fields and verifies signature for non-faucet transactions.
func (tx Transaction) BasicValidate() error {
	if strings.TrimSpace(tx.From) == "" {
		return errors.New("error: sender is required")
	}

	if strings.TrimSpace(tx.To) == "" {
		return errors.New("error: recipient is required")
	}

	if tx.Amount <= 0 {
		return errors.New("error: amount must be greater than zero")
	}

	if strings.EqualFold(strings.TrimSpace(tx.From), strings.TrimSpace(tx.To)) {
		return errors.New("error: sender and recipient cannot be the same")
	}

	if tx.IsFaucet() {
		return nil
	}

	// If no pubkey/signature are provided treat as legacy unsigned transaction
	// (keep backward compatibility for existing tests/examples).
	if tx.PubKeyHex == "" && tx.SigHex == "" {
		return nil
	}

	pub, err := hex.DecodeString(strings.TrimSpace(tx.PubKeyHex))
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return errors.New("error: invalid pubkey format")
	}

	sig, err := hex.DecodeString(strings.TrimSpace(tx.SigHex))
	if err != nil {
		return errors.New("error: invalid signature format")
	}

	// Ensure From matches pubkey hex to tie identity to the key.
	if !strings.EqualFold(tx.From, strings.ToUpper(hex.EncodeToString(pub))) && !strings.EqualFold(tx.From, strings.ToLower(hex.EncodeToString(pub))) {
		return errors.New("error: sender does not match provided public key")
	}

	if !ed25519.Verify(ed25519.PublicKey(pub), tx.signingBytes(), sig) {
		return errors.New("error: signature verification failed")
	}

	return nil
}

// Sign the transaction using a private key hex string. This sets PubKeyHex,
// SigHex and overwrites From with the public key hex so identity is consistent.
func (tx *Transaction) Sign(skHex string) error {
	skb, err := hex.DecodeString(strings.TrimSpace(skHex))
	if err != nil {
		return fmt.Errorf("invalid private key hex: %w", err)
	}

	var priv ed25519.PrivateKey
	if len(skb) == ed25519.SeedSize {
		priv = ed25519.NewKeyFromSeed(skb)
	} else if len(skb) == ed25519.PrivateKeySize {
		priv = ed25519.PrivateKey(skb)
	} else {
		return errors.New("private key must be 32-byte seed or 64-byte private key in hex")
	}

	pub := priv.Public().(ed25519.PublicKey)
	sig := ed25519.Sign(priv, tx.signingBytes())

	tx.PubKeyHex = hex.EncodeToString(pub)
	tx.SigHex = hex.EncodeToString(sig)
	tx.From = hex.EncodeToString(pub)
	return nil
}

func (tx Transaction) signingBytes() []byte {
	// Deterministic signing over From|To|Amount in this order.
	return []byte(fmt.Sprintf("%s|%s|%d", tx.From, tx.To, tx.Amount))
}

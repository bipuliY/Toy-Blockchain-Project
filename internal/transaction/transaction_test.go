package transaction

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestValidateNetworkRejectsUnsignedTransaction(t *testing.T) {
	tx := New(
		"Alice",
		"Bob",
		10,
	)

	err := tx.ValidateNetwork()

	if err == nil {
		t.Fatal(
			"expected unsigned network transaction to be rejected",
		)
	}
}

func TestValidateNetworkAcceptsValidSignedTransaction(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf(
			"failed to generate key pair: %v",
			err,
		)
	}

	publicKeyHex := hex.EncodeToString(publicKey)
	privateKeyHex := hex.EncodeToString(privateKey)

	tx := New(
		publicKeyHex,
		"Bob",
		10,
	)

	if err := tx.Sign(privateKeyHex); err != nil {
		t.Fatalf(
			"failed to sign transaction: %v",
			err,
		)
	}

	if err := tx.ValidateNetwork(); err != nil {
		t.Fatalf(
			"expected valid signed transaction, got error: %v",
			err,
		)
	}
}
func TestTransactionIDIsDeterministic(t *testing.T) {
	tx := New(
		"Alice",
		"Bob",
		10,
	)

	first := tx.ID()
	second := tx.ID()

	if first == "" {
		t.Fatal("expected transaction ID")
	}

	if first != second {
		t.Fatalf(
			"expected identical IDs, got %s and %s",
			first,
			second,
		)
	}
}
func TestTransactionIDChangesWhenTransactionChanges(t *testing.T) {
	tx := New(
		"Alice",
		"Bob",
		10,
	)

	originalID := tx.ID()

	tx.Amount = 20

	changedID := tx.ID()

	if originalID == changedID {
		t.Fatal(
			"expected transaction ID to change after transaction changed",
		)
	}
}

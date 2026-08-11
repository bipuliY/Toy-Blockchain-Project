package node

import (
	"testing"

	"toy-blockchain/chain"
)

func TestNewNode(t *testing.T) {
	n := New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	if n.Address() != "localhost:8001" {
		t.Fatalf(
			"expected address localhost:8001, got %s",
			n.Address(),
		)
	}

	if n.Height() != 0 {
		t.Fatalf(
			"expected genesis height 0, got %d",
			n.Height(),
		)
	}

	if n.HeadHash() == "" {
		t.Fatal("expected genesis block head hash")
	}

	if len(n.Peers()) != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			len(n.Peers()),
		)
	}
}
func TestNodeStatus(t *testing.T) {
	n := New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	status := n.Status()

	if status.Address != "localhost:8001" {
		t.Fatalf(
			"expected address localhost:8001, got %s",
			status.Address,
		)
	}

	if status.Height != 0 {
		t.Fatalf(
			"expected height 0, got %d",
			status.Height,
		)
	}

	if status.HeadHash == "" {
		t.Fatal("expected genesis head hash")
	}

	if status.PeerCount != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			status.PeerCount,
		)
	}
}
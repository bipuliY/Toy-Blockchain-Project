package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest" //this let us test HTTP handler without actually starting port 8001
	"strings"
	"testing"

	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"toy-blockchain/block"
	"toy-blockchain/internal/transaction"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/node"
)

func TestStatusEndpoint(t *testing.T) {
	n := node.New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	server := api.NewServer(n)

	request := httptest.NewRequest(
		http.MethodGet,
		"/status",
		nil,
	)

	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var status node.Status

	if err := json.NewDecoder(recorder.Body).Decode(&status); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

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
		t.Fatal("expected non-empty head hash")
	}

	if status.PeerCount != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			status.PeerCount,
		)
	}
}
func TestTransactionEndpointRejectsUnsignedTransaction(
	t *testing.T,
) {
	n := node.New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	server := api.NewServer(n)

	body := strings.NewReader(`
		{
			"from": "Alice",
			"to": "Bob",
			"amount": 10
		}
	`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		body,
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	server.Handler().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	if n.PendingCount() != 0 {
		t.Fatalf(
			"expected no pending transactions, got %d",
			n.PendingCount(),
		)
	}
}
func TestTransactionEndpointAcceptsSignedTransaction(t *testing.T) {
	n := node.New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	sender := hex.EncodeToString(publicKey)

	// Give the sender funds first.
	fundingTx := transaction.New(
		transaction.Faucet,
		sender,
		100,
	)

	if err := n.SubmitTransaction(fundingTx); err != nil {
		t.Fatalf("failed to fund sender: %v", err)
	}

	// Create the real transaction.
	tx := transaction.New(
		sender,
		"Bob",
		25,
	)

	if err := tx.Sign(
		hex.EncodeToString(privateKey),
	); err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	body, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("failed to encode transaction: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/transactions",
		strings.NewReader(string(body)),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	api.NewServer(n).
		Handler().
		ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if n.PendingCount() != 2 {
		t.Fatalf(
			"expected 2 pending transactions, got %d",
			n.PendingCount(),
		)
	}
}
func TestMineEndpoint(t *testing.T) {
	n := node.New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/mine",
		nil,
	)

	recorder := httptest.NewRecorder()

	api.NewServer(n).
		Handler().
		ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if n.Height() != 1 {
		t.Fatalf(
			"expected height 1, got %d",
			n.Height(),
		)
	}

	if n.PendingCount() != 0 {
		t.Fatalf(
			"expected 0 pending transactions, got %d",
			n.PendingCount(),
		)
	}
}
func TestMineEndpointGossipsBlockToPeer(
	t *testing.T,
) {
	nodeB := node.New(
		"node-b",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	serverB := httptest.NewServer(
		api.NewServer(nodeB).Handler(),
	)
	defer serverB.Close()

	nodeA := node.New(
		"node-a",
		[]string{
			serverB.URL,
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := nodeA.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/mine",
		nil,
	)

	recorder := httptest.NewRecorder()

	api.NewServer(nodeA).
		Handler().
		ServeHTTP(
			recorder,
			request,
		)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if nodeA.Height() != 1 {
		t.Fatalf(
			"expected node A height 1, got %d",
			nodeA.Height(),
		)
	}

	if nodeB.Height() != 1 {
		t.Fatalf(
			"expected node B height 1, got %d",
			nodeB.Height(),
		)
	}

	if nodeA.HeadHash() != nodeB.HeadHash() {
		t.Fatalf(
			"expected matching heads: A=%s B=%s",
			nodeA.HeadHash(),
			nodeB.HeadHash(),
		)
	}
}
func TestGetBlocksEndpoint(
	t *testing.T,
) {
	n := node.New(
		"node-a",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"failed to submit transaction: %v",
			err,
		)
	}

	minedBlock, _, err := n.MinePending()
	if err != nil {
		t.Fatalf(
			"failed to mine block: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/blocks?from=1",
		nil,
	)

	recorder := httptest.NewRecorder()

	api.NewServer(n).
		Handler().
		ServeHTTP(
			recorder,
			request,
		)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		From   int           `json:"from"`
		Count  int           `json:"count"`
		Blocks []block.Block `json:"blocks"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.From != 1 {
		t.Fatalf(
			"expected from 1, got %d",
			response.From,
		)
	}

	if response.Count != 1 {
		t.Fatalf(
			"expected 1 block, got %d",
			response.Count,
		)
	}

	if len(response.Blocks) != 1 {
		t.Fatalf(
			"expected 1 returned block, got %d",
			len(response.Blocks),
		)
	}

	if response.Blocks[0].Hash != minedBlock.Hash {
		t.Fatalf(
			"expected hash %s, got %s",
			minedBlock.Hash,
			response.Blocks[0].Hash,
		)
	}
}
func TestBalancesEndpoint(
	t *testing.T,
) {
	n := node.New(
		"localhost:8001",
		nil,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	tx := transaction.New(
		transaction.Faucet,
		"Alice",
		100,
	)

	if err := n.SubmitTransaction(tx); err != nil {
		t.Fatalf(
			"submit transaction failed: %v",
			err,
		)
	}

	if _, _, err := n.MinePending(); err != nil {
		t.Fatalf(
			"mine transaction failed: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/balances",
		nil,
	)

	recorder := httptest.NewRecorder()

	api.NewServer(n).
		Handler().
		ServeHTTP(
			recorder,
			request,
		)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response struct {
		Balances map[string]int `json:"balances"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if response.Balances["Alice"] != 100 {
		t.Fatalf(
			"expected Alice balance 100, got %d",
			response.Balances["Alice"],
		)
	}
}
func TestPeersEndpoint(
	t *testing.T,
) {
	n := node.New(
		"localhost:8001",
		[]string{
			"http://localhost:8002",
			"http://localhost:8003",
		},
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/peers",
		nil,
	)

	recorder := httptest.NewRecorder()

	api.NewServer(n).
		Handler().
		ServeHTTP(
			recorder,
			request,
		)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response struct {
		Peers []string `json:"peers"`
	}

	if err := json.NewDecoder(
		recorder.Body,
	).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v",
			err,
		)
	}

	if len(response.Peers) != 2 {
		t.Fatalf(
			"expected 2 peers, got %d",
			len(response.Peers),
		)
	}
}

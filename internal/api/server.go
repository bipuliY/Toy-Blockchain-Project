package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"toy-blockchain/internal/network"

	"toy-blockchain/internal/node"
	"toy-blockchain/internal/transaction"
)

// Server exposes a blockchain node through HTTP.
type Server struct {
	node       *node.Node
	peerClient *network.Client
}

// NewServer creates an HTTP server wrapper around a node.
func NewServer(n *node.Node) *Server {
	return &Server{
		node:       n,
		peerClient: network.NewClient(),
	}
}

// Handler returns the HTTP handler used by the node server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /status", s.handleStatus)
	mux.HandleFunc(
		"POST /transactions",
		s.handleTransaction,
	)

	return mux
}

// handleStatus returns the node's current status as JSON.
func (s *Server) handleStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	status := s.node.Status()

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(
			w,
			"failed to encode status",
			http.StatusInternalServerError,
		)
		return
	}
}

// handleTransaction accepts a transaction and adds it
// to the node's pending pool if it is valid.
func (s *Server) handleTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	var tx transaction.Transaction

	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(
			w,
			"invalid transaction JSON",
			http.StatusBadRequest,
		)
		return
	}

	// if err := s.node.SubmitTransaction(tx); err != nil {
	// 	http.Error(
	// 		w,
	// 		err.Error(),
	// 		http.StatusBadRequest,
	// 	)
	// 	return
	// }
	if err := s.node.SubmitTransaction(tx); err != nil {
		if errors.Is(
			err,
			node.ErrTransactionAlreadySeen,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_ = json.NewEncoder(w).Encode(
				map[string]any{
					"status": "duplicate",
				},
			)

			return
		}

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)

		return
	}
	for _, peer := range s.node.Peers() {
		if err := s.peerClient.SendTransaction(
			r.Context(),
			peer,
			tx,
		); err != nil {
			log.Printf(
				"transaction gossip failed peer=%s tx=%s error=%v",
				peer,
				tx.ID(),
				err,
			)
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status       string `json:"status"`
		PendingCount int    `json:"pending_count"`
	}{
		Status:       "accepted",
		PendingCount: s.node.PendingCount(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

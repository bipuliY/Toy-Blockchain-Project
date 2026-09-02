package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"toy-blockchain/block"
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

	mux.HandleFunc(
		"POST /mine",
		s.handleMine,
	)

	mux.HandleFunc(
		"GET /status",
		s.handleStatus,
	)
	mux.HandleFunc(
		"GET /balances",
		s.handleBalances,
	)
	mux.HandleFunc(
		"GET /peers",
		s.handlePeers,
	)

	mux.HandleFunc(
		"POST /transactions",
		s.handleTransaction,
	)

	mux.HandleFunc(
		"POST /blocks",
		s.handleBlock,
	)
	mux.HandleFunc(
		"GET /blocks",
		s.handleGetBlocks,
	)

	return mux
}
func (s *Server) handleBlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	var receivedBlock block.Block

	if err := json.NewDecoder(r.Body).Decode(
		&receivedBlock,
	); err != nil {
		http.Error(
			w,
			"invalid block JSON",
			http.StatusBadRequest,
		)
		return
	}

	if err := s.node.AcceptBlock(
		receivedBlock,
	); err != nil {

		if errors.Is(
			err,
			node.ErrBlockAlreadySeen,
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

	// Forward a newly accepted block to peers.
	for _, peer := range s.node.Peers() {
		if err := s.peerClient.SendBlock(
			r.Context(),
			peer,
			receivedBlock,
		); err != nil {
			log.Printf(
				"block gossip failed peer=%s block=%s error=%v",
				peer,
				receivedBlock.Hash,
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
		Status   string `json:"status"`
		Height   int    `json:"height"`
		HeadHash string `json:"head_hash"`
	}{
		Status:   "accepted",
		Height:   s.node.Height(),
		HeadHash: s.node.HeadHash(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

// handleGetBlocks returns blockchain blocks starting
// from the requested height.
func (s *Server) handleGetBlocks(
	w http.ResponseWriter,
	r *http.Request,
) {
	fromValue := r.URL.Query().Get("from")

	if fromValue == "" {
		http.Error(
			w,
			"from query parameter is required",
			http.StatusBadRequest,
		)
		return
	}

	fromHeight, err := strconv.Atoi(fromValue)
	if err != nil {
		http.Error(
			w,
			"from must be an integer",
			http.StatusBadRequest,
		)
		return
	}

	blocks, err := s.node.BlocksFrom(
		fromHeight,
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	response := struct {
		From   int           `json:"from"`
		Count  int           `json:"count"`
		Blocks []block.Block `json:"blocks"`
	}{
		From:   fromHeight,
		Count:  len(blocks),
		Blocks: blocks,
	}

	if err := json.NewEncoder(w).Encode(
		response,
	); err != nil {
		http.Error(
			w,
			"failed to encode blocks",
			http.StatusInternalServerError,
		)
		return
	}
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

// handleMine mines the node's pending transactions.
func (s *Server) handleMine(
	w http.ResponseWriter,
	r *http.Request,
) {
	minedBlock, mineResult, err := s.node.MinePending()
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	// Automatically gossip the newly mined block.
	for _, peer := range s.node.Peers() {
		if err := s.peerClient.SendBlock(
			r.Context(),
			peer,
			minedBlock,
		); err != nil {
			log.Printf(
				"block gossip failed peer=%s block=%s error=%v",
				peer,
				minedBlock.Hash,
				err,
			)
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	response := struct {
		Status     string           `json:"status"`
		Block      block.Block      `json:"block"`
		MineResult block.MineResult `json:"mine_result"`
	}{
		Status:     "mined",
		Block:      minedBlock,
		MineResult: mineResult,
	}

	if err := json.NewEncoder(w).Encode(
		response,
	); err != nil {
		return
	}
}

// func (s *Server) handleMine(
// 	w http.ResponseWriter,
// 	r *http.Request,
// ) {
// 	minedBlock, mineResult, err := s.node.MinePending()
// 	if err != nil {
// 		http.Error(
// 			w,
// 			err.Error(),
// 			http.StatusBadRequest,
// 		)
// 		return
// 	}
// 	for _, peer := range s.node.Peers() {
// 		if err := s.peerClient.SendBlock(
// 			r.Context(),
// 			peer,
// 			minedBlock,
// 		); err != nil {
// 			log.Printf(
// 				"block gossip failed peer=%s block=%s error=%v",
// 				peer,
// 				minedBlock.Hash,
// 				err,
// 			)
// 		}
// 	}
// 	w.Header().Set(
// 		"Content-Type",
// 		"application/json",
// 	)

// 	response := struct {
// 		Status     string           `json:"status"`
// 		Block      block.Block      `json:"block"`
// 		MineResult block.MineResult `json:"mine_result"`
// 	}{
// 		Status:     "mined",
// 		Block:      minedBlock,
// 		MineResult: mineResult,
// 	}

//		if err := json.NewEncoder(w).Encode(response); err != nil {
//			return
//		}
//	}
//
// handleBalances returns confirmed blockchain balances.
func (s *Server) handleBalances(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	response := struct {
		Balances map[string]int `json:"balances"`
	}{
		Balances: s.node.Balances(),
	}

	if err := json.NewEncoder(w).Encode(
		response,
	); err != nil {
		http.Error(
			w,
			"failed to encode balances",
			http.StatusInternalServerError,
		)
		return
	}
}

// handlePeers returns the peers currently known
// by this blockchain node.
func (s *Server) handlePeers(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	response := struct {
		Peers []string `json:"peers"`
	}{
		Peers: s.node.Peers(),
	}

	if err := json.NewEncoder(w).Encode(
		response,
	); err != nil {
		http.Error(
			w,
			"failed to encode peers",
			http.StatusInternalServerError,
		)
		return
	}
}

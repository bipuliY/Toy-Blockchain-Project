package api

import (
	"encoding/json"
	"net/http"

	"toy-blockchain/internal/node"
)

// Server exposes a blockchain node through HTTP.
type Server struct {
	node *node.Node
}

// NewServer creates an HTTP server wrapper around a node.
func NewServer(n *node.Node) *Server {
	return &Server{
		node: n,
	}
}

// Handler returns the HTTP handler used by the node server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /status", s.handleStatus)

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
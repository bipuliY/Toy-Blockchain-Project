package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/node"
)

func main() {
	address := flag.String(
		"address",
		"localhost:8001",
		"HTTP address the node listens on",
	)

	peersFlag := flag.String(
		"peers",
		"",
		"comma-separated peer URLs",
	)

	flag.Parse()

	peers := parsePeers(*peersFlag)

	n := node.New(
		*address,
		peers,
		chain.DefaultDifficulty,
		chain.DefaultBlockSize,
	)

	apiServer := api.NewServer(n)

	server := &http.Server{
		Addr:              *address,
		Handler:           apiServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"node starting address=%s peers=%d",
		*address,
		len(peers),
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("node server failed: %v", err)
	}
}

func parsePeers(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	rawPeers := strings.Split(value, ",")
	peers := make([]string, 0, len(rawPeers))

	for _, peer := range rawPeers {
		peer = strings.TrimSpace(peer)

		if peer == "" {
			continue
		}

		peers = append(peers, peer)
	}

	return peers
}
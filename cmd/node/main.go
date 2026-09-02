package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
	"toy-blockchain/internal/syncer"

	"toy-blockchain/chain"
	"toy-blockchain/internal/api"
	"toy-blockchain/internal/network"
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
	peerClient := network.NewClient()

	
	for _, peer := range n.Peers() {
		if err := syncer.SyncFromPeer(
			context.Background(),
			n,
			peerClient,
			peer,
		); err != nil {
			log.Printf(
				"sync failed peer=%s error=%v",
				peer,
				err,
			)
			continue
		}

		log.Printf(
			"sync complete peer=%s local_height=%d head=%s",
			peer,
			n.Height(),
			n.HeadHash(),
		)
	}
	// Periodically check peers so a running node can
	// catch up if it falls behind.
	go func() {
		ticker := time.NewTicker(
			5 * time.Second,
		)
		defer ticker.Stop()

		for range ticker.C {
			for _, peer := range n.Peers() {
				if err := syncer.SyncFromPeer(
					context.Background(),
					n,
					peerClient,
					peer,
				); err != nil {
					log.Printf(
						"periodic sync failed peer=%s error=%v",
						peer,
						err,
					)
					continue
				}
			}
		}
	}()
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

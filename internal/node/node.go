package node

import (
	"strings"
	"sync"

	"toy-blockchain/chain"
)

// Node represents one running blockchain node.
//
// Later this type will also coordinate:
//   - HTTP requests
//   - transaction gossip
//   - block gossip
//   - chain synchronisation
//   - fork resolution
//
// For now it only owns the blockchain and basic network configuration.
type Node struct {
	mu sync.RWMutex  //protect shared state 

	blockchain *chain.Blockchain
	address    string
	peers      map[string]struct{}
}

// New creates a new blockchain node.
//
// address is the HTTP address this node will eventually listen on.
// peers contains the initial addresses of other nodes it knows about.
func New(
	address string,
	peers []string,
	difficulty int,
	blockSize int,
) *Node {
	peerSet := make(map[string]struct{})

	for _, peer := range peers {
		peer = strings.TrimSpace(peer)

		if peer == "" {
			continue
		}

		peerSet[peer] = struct{}{}
	}

	return &Node{
		blockchain: chain.NewBlockchain(difficulty, blockSize),
		address:    strings.TrimSpace(address),
		peers:      peerSet,
	}
}

// Address returns the configured address of this node.
func (n *Node) Address() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.address
}

// Height returns the height of the current blockchain.
//
// The genesis block has height 0.
func (n *Node) Height() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.blockchain.Blocks) == 0 {
		return -1
	}

	return len(n.blockchain.Blocks) - 1
}

// HeadHash returns the hash of the most recent block.
func (n *Node) HeadHash() string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(n.blockchain.Blocks) == 0 {
		return ""
	}

	return n.blockchain.Blocks[len(n.blockchain.Blocks)-1].Hash
}

// Peers returns a copy of the node's current peer list.
//
// Returning a copy prevents callers from modifying the internal peer map.
func (n *Node) Peers() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]string, 0, len(n.peers))

	for peer := range n.peers {
		peers = append(peers, peer)
	}

	return peers
}
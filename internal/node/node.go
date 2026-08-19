package node

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"toy-blockchain/block"
	"toy-blockchain/chain"
	"toy-blockchain/internal/transaction"
)

var ErrTransactionAlreadySeen = errors.New(
	"transaction already seen",
)
var ErrBlockAlreadySeen = errors.New(
	"block already seen",
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
	mu sync.RWMutex //protect shared state

	blockchain       *chain.Blockchain // the node's copy of the blockchain
	address          string
	peers            map[string]struct{}
	seenTransactions map[string]struct{}
	seenBlocks       map[string]struct{}
}

// Status represents a read-only snapshot of the node.
type Status struct {
	Address      string `json:"address"`
	Height       int    `json:"height"`
	HeadHash     string `json:"head_hash"`
	PeerCount    int    `json:"peer_count"`
	PendingCount int    `json:"pending_count"`
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
		blockchain:       chain.NewBlockchain(difficulty, blockSize),
		address:          strings.TrimSpace(address),
		peers:            peerSet,
		seenTransactions: make(map[string]struct{}),
		seenBlocks:       make(map[string]struct{}),
	}
}

// MinePending mines pending transactions into a new block.
// func (n *Node) MinePending() (
// 	block.Block,
// 	block.MineResult,
// 	error,
// ) {
// 	n.mu.Lock()
// 	defer n.mu.Unlock()

// 	// return n.blockchain.MinePending()
// 	minedBlock, result, err :=
// 	n.blockchain.MinePending()

// if err != nil {
// 	return block.Block{}, block.MineResult{}, err
// }

// n.seenBlocks[minedBlock.Hash] = struct{}{}

// return minedBlock, result, nil
// }
func (n *Node) MinePending() (
	block.Block,
	block.MineResult,
	error,
) {
	n.mu.Lock()
	defer n.mu.Unlock()

	minedBlock, result, err :=
		n.blockchain.MinePending()

	if err != nil {
		return block.Block{}, block.MineResult{}, err
	}

	// This node mined the block itself,
	// so it has already seen this block.
	n.seenBlocks[minedBlock.Hash] = struct{}{}

	return minedBlock, result, nil
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

// BlocksFrom returns a safe copy of blockchain blocks
// starting at the given height.
func (n *Node) BlocksFrom(
	fromHeight int,
) ([]block.Block, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if fromHeight < 0 {
		return nil, errors.New(
			"block height cannot be negative",
		)
	}

	if fromHeight > len(n.blockchain.Blocks) {
		return nil, fmt.Errorf(
			"block height %d is beyond current chain",
			fromHeight,
		)
	}

	source := n.blockchain.Blocks[fromHeight:]

	blocks := make(
		[]block.Block,
		len(source),
	)

	copy(blocks, source)

	// Each block contains a transaction slice.
	// Copy that slice too so callers cannot modify
	// the node's internal blockchain data.
	for i := range blocks {
		blocks[i].Transactions = append(
			[]transaction.Transaction(nil),
			blocks[i].Transactions...,
		)
	}

	return blocks, nil
}

// ChainSnapshot returns a safe copy of the entire
// blockchain owned by this node.
func (n *Node) ChainSnapshot() []block.Block {
	n.mu.RLock()
	defer n.mu.RUnlock()

	blocks := make(
		[]block.Block,
		len(n.blockchain.Blocks),
	)

	copy(
		blocks,
		n.blockchain.Blocks,
	)

	// Each block contains its own transaction slice.
	// Copy that slice too so callers cannot modify
	// the node's internal blockchain data.
	for i := range blocks {
		blocks[i].Transactions = append(
			[]transaction.Transaction(nil),
			blocks[i].Transactions...,
		)
	}

	return blocks
}

// ValidateCandidateChain checks whether a complete chain
// received from a peer is valid without changing this node.
func (n *Node) ValidateCandidateChain(
	blocks []block.Block,
) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if len(blocks) == 0 {
		return errors.New(
			"candidate chain has no blocks",
		)
	}

	if len(n.blockchain.Blocks) == 0 {
		return errors.New(
			"local blockchain has no genesis block",
		)
	}

	// Both chains must belong to the same blockchain.
	if blocks[0].Hash != n.blockchain.Blocks[0].Hash {
		return errors.New(
			"candidate chain has different genesis block",
		)
	}

	// A genesis-only candidate is valid if it matches
	// our known genesis block.
	if len(blocks) == 1 {
		return nil
	}

	// Network blocks must contain valid signed
	// network transactions.
	for _, b := range blocks[1:] {
		for _, tx := range b.Transactions {
			if err := tx.ValidateNetwork(); err != nil {
				return fmt.Errorf(
					"invalid transaction in candidate block %d: %w",
					b.Height,
					err,
				)
			}
		}
	}

	// Make a deep copy so validation cannot affect
	// the data supplied by the caller.
	candidateBlocks := make(
		[]block.Block,
		len(blocks),
	)

	copy(
		candidateBlocks,
		blocks,
	)

	for i := range candidateBlocks {
		candidateBlocks[i].Transactions = append(
			[]transaction.Transaction(nil),
			candidateBlocks[i].Transactions...,
		)
	}

	// Use the same blockchain configuration as this node.
	candidate := &chain.Blockchain{
		Blocks:                 candidateBlocks,
		PendingTransactions:    nil,
		Difficulty:             candidateBlocks[len(candidateBlocks)-1].Difficulty,
		BlockSize:              n.blockchain.BlockSize,
		TargetBlockTimeSeconds: n.blockchain.TargetBlockTimeSeconds,
		RetargetInterval:       n.blockchain.RetargetInterval,
		MinDifficulty:          n.blockchain.MinDifficulty,
		MaxDifficulty:          n.blockchain.MaxDifficulty,
	}

	// Difficulty stores the difficulty expected for
	// the NEXT block.
	candidate.Difficulty =
		candidate.CalculateNextDifficulty()

	validation := candidate.Validate()

	if !validation.Valid {
		return fmt.Errorf(
			"invalid candidate chain at block %d: %s",
			validation.BlockHeight,
			validation.Reason,
		)
	}

	return nil
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

// Status returns a consistent snapshot of the node's current state.
func (n *Node) Status() Status {
	n.mu.RLock()
	defer n.mu.RUnlock()

	height := -1
	headHash := ""

	if len(n.blockchain.Blocks) > 0 {
		height = len(n.blockchain.Blocks) - 1
		headHash = n.blockchain.Blocks[len(n.blockchain.Blocks)-1].Hash
	}

	return Status{
		Address:      n.address,
		Height:       height,
		HeadHash:     headHash,
		PeerCount:    len(n.peers),
		PendingCount: len(n.blockchain.PendingTransactions),
	}
}

// SubmitTransaction validates a network transaction and,
// if valid, adds it to this node's pending transaction pool.
// func (n *Node) SubmitTransaction(
// 	tx transaction.Transaction,
// ) error {
// 	if err := tx.ValidateNetwork(); err != nil {
// 		return err
// 	}

// 	n.mu.Lock()
// 	defer n.mu.Unlock()

//		return n.blockchain.AddTransaction(tx)
//	}
func (n *Node) SubmitTransaction(
	tx transaction.Transaction,
) error {
	if err := tx.ValidateNetwork(); err != nil {
		return err
	}

	txID := tx.ID()

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.seenTransactions[txID]; exists {
		return ErrTransactionAlreadySeen
	}

	if err := n.blockchain.AddTransaction(tx); err != nil {
		return err
	}

	n.seenTransactions[txID] = struct{}{}

	return nil
}

// PendingCount returns the number of transactions waiting to be mined.
func (n *Node) PendingCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return len(n.blockchain.PendingTransactions)
}

// AcceptBlock validates a block received from a peer and,
// if valid, appends it to the current blockchain.
func (n *Node) AcceptBlock(
	b block.Block,
) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, exists := n.seenBlocks[b.Hash]; exists {
		return ErrBlockAlreadySeen
	}
	if len(n.blockchain.Blocks) == 0 {
		return errors.New(
			"blockchain has no genesis block",
		)
	}

	expectedHeight := len(n.blockchain.Blocks)

	if b.Height != expectedHeight {
		return fmt.Errorf(
			"unexpected block height: expected %d, got %d",
			expectedHeight,
			b.Height,
		)
	}

	currentHead :=
		n.blockchain.Blocks[len(n.blockchain.Blocks)-1]

	if b.PrevHash != currentHead.Hash {
		return errors.New(
			"block does not extend current head",
		)
	}

	// Network blocks must contain valid network transactions.
	for _, tx := range b.Transactions {
		if err := tx.ValidateNetwork(); err != nil {
			return fmt.Errorf(
				"invalid block transaction: %w",
				err,
			)
		}
	}

	// Validate using a temporary candidate chain first.
	// Do not modify the real blockchain until validation passes.
	candidate := *n.blockchain

	candidate.Blocks = append(
		[]block.Block(nil),
		n.blockchain.Blocks...,
	)

	candidate.Blocks = append(
		candidate.Blocks,
		b,
	)

	// Store the difficulty expected for the block after this one.
	candidate.Difficulty =
		candidate.CalculateNextDifficulty()

	validation := candidate.Validate()

	if !validation.Valid {
		return fmt.Errorf(
			"invalid block: %s",
			validation.Reason,
		)
	}

	// Find transactions that are now confirmed in this block.
	confirmed := make(
		map[string]struct{},
		len(b.Transactions),
	)

	for _, tx := range b.Transactions {
		confirmed[tx.ID()] = struct{}{}
	}

	// Rebuild the pending pool.
	oldPending := append(
		[]transaction.Transaction(nil),
		n.blockchain.PendingTransactions...,
	)

	candidate.PendingTransactions = nil

	for _, tx := range oldPending {
		if _, exists := confirmed[tx.ID()]; exists {
			continue
		}

		// Keep only transactions that are still valid after
		// accepting the new block.
		if err := candidate.AddTransaction(tx); err != nil {
			continue
		}
	}

	// Everything passed. Now replace the real chain.
	n.blockchain = &candidate

	// Transactions confirmed by a block should also remain
	// known to this node, so they cannot be submitted again.
	for _, tx := range b.Transactions {
		n.seenTransactions[tx.ID()] = struct{}{}
	}

	n.seenBlocks[b.Hash] = struct{}{}

	return nil
}

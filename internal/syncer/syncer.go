package syncer

import (
	"context"
	"errors"
	"fmt"

	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
)

var ErrForkDetected = errors.New(
	"fork detected",
)

// SyncFromPeer downloads and validates missing blocks
// from one peer when that peer is ahead.
func SyncFromPeer(
	ctx context.Context,
	n *node.Node,
	client *network.Client,
	peerURL string,
) error {
	peerStatus, err := client.FetchStatus(
		ctx,
		peerURL,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch peer status: %w",
			err,
		)
	}

	// Take one consistent snapshot of our local status.
	localStatus := n.Status()

	// Peer is shorter.
	if peerStatus.Height < localStatus.Height {
		return nil
	}

	// Same height.
	if peerStatus.Height == localStatus.Height {
		// Same height + same hash means both nodes agree.
		if peerStatus.HeadHash == localStatus.HeadHash {
			return nil
		}

		// Same height + different hash means fork.
		return fmt.Errorf(
			"%w: height=%d local_head=%s peer_head=%s",
			ErrForkDetected,
			localStatus.Height,
			localStatus.HeadHash,
			peerStatus.HeadHash,
		)
	}

	// Peer is ahead.
	//
	// Start from our CURRENT height so we can compare
	// our current head with the peer's block at that height.
	blocks, err := client.FetchBlocks(
		ctx,
		peerURL,
		localStatus.Height,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch peer blocks: %w",
			err,
		)
	}

	if len(blocks) == 0 {
		return fmt.Errorf(
			"peer returned no blocks from height %d",
			localStatus.Height,
		)
	}

	if blocks[0].Height != localStatus.Height {
		return fmt.Errorf(
			"unexpected peer block height: expected %d, got %d",
			localStatus.Height,
			blocks[0].Height,
		)
	}

	// Peer is ahead, but its chain has already diverged
	// from ours.
	// The peer is ahead, but its chain is different
	// from ours at our current height.
	//
	// Because the peer has the longer chain, download
	// its complete chain and try a reorganization.
	if blocks[0].Hash != localStatus.HeadHash {
		fullChain, err := client.FetchBlocks(
			ctx,
			peerURL,
			0,
		)
		if err != nil {
			return fmt.Errorf(
				"fetch full peer chain for reorg: %w",
				err,
			)
		}

		if err := n.AdoptCandidateChain(
			fullChain,
		); err != nil {
			return fmt.Errorf(
				"adopt longer peer chain: %w",
				err,
			)
		}

		return nil
	}

	// blocks[0] is our existing current head.
	// Accept only the newer blocks after it.
	for _, b := range blocks[1:] {
		if err := n.AcceptBlock(b); err != nil {
			return fmt.Errorf(
				"accept synced block height %d: %w",
				b.Height,
				err,
			)
		}
	}

	if n.Height() < peerStatus.Height {
		return fmt.Errorf(
			"sync incomplete: local height %d, peer height %d",
			n.Height(),
			peerStatus.Height,
		)
	}

	return nil
}

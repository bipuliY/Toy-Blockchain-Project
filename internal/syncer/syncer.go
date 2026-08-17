package syncer

import (
	"context"
	"fmt"

	"toy-blockchain/internal/network"
	"toy-blockchain/internal/node"
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

	localHeight := n.Height()

	// Nothing to do if this node is already caught up.
	if peerStatus.Height <= localHeight {
		return nil
	}

	fromHeight := localHeight + 1

	blocks, err := client.FetchBlocks(
		ctx,
		peerURL,
		fromHeight,
	)
	if err != nil {
		return fmt.Errorf(
			"fetch missing blocks: %w",
			err,
		)
	}

	// Validate and append every downloaded block
	// using the existing Node validation logic.
	for _, b := range blocks {
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

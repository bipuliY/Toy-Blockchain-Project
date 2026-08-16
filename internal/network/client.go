package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toy-blockchain/block"

	"toy-blockchain/internal/node"
	"toy-blockchain/internal/transaction"
)

// Client sends HTTP requests to peer blockchain nodes.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a peer HTTP client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // set a timeout for HTTP requests
		},
	}
}

// FetchStatus requests the current status of a peer node.
func (c *Client) FetchStatus(
	ctx context.Context,
	peerURL string,
) (node.Status, error) {
	peerURL = strings.TrimRight(
		strings.TrimSpace(peerURL),
		"/",
	)

	if peerURL == "" {
		return node.Status{}, fmt.Errorf(
			"peer URL is required",
		)
	}

	url := peerURL + "/status"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return node.Status{}, fmt.Errorf(
			"create status request: %w",
			err,
		)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return node.Status{}, fmt.Errorf(
			"request peer status: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return node.Status{}, fmt.Errorf(
			"peer returned status %d",
			resp.StatusCode,
		)
	}

	var status node.Status

	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return node.Status{}, fmt.Errorf(
			"decode peer status: %w",
			err,
		)
	}

	return status, nil
}

// SendTransaction sends a transaction to a peer node.
func (c *Client) SendTransaction(
	ctx context.Context,
	peerURL string,
	tx transaction.Transaction,
) error {
	peerURL = strings.TrimRight(
		strings.TrimSpace(peerURL),
		"/",
	)

	if peerURL == "" {
		return fmt.Errorf(
			"peer URL is required",
		)
	}

	body, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf(
			"encode transaction: %w",
			err,
		)
	}

	url := peerURL + "/transactions"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf(
			"create transaction request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"send transaction to peer: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"peer rejected transaction with status %d",
			resp.StatusCode,
		)
	}

	return nil
}

// SendBlock sends a mined block to a peer node.
func (c *Client) SendBlock(
	ctx context.Context,
	peerURL string,
	b block.Block,
) error {
	peerURL = strings.TrimRight(
		strings.TrimSpace(peerURL),
		"/",
	)

	if peerURL == "" {
		return fmt.Errorf(
			"peer URL is required",
		)
	}

	body, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf(
			"encode block: %w",
			err,
		)
	}

	url := peerURL + "/blocks"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf(
			"create block request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"send block to peer: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"peer rejected block with status %d",
			resp.StatusCode,
		)
	}

	return nil
}

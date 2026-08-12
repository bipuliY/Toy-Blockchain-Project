package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toy-blockchain/internal/node"
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

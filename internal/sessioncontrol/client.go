package sessioncontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/DashSaman/PV-NaivePanel/internal/sessionkill"
)

// Client is an API-side HTTP-over-Unix-socket client that sends kill
// commands to the overlay's session-control listener. It implements the
// sessionkill.SessionKiller interface so the HTTP handler can remain
// decoupled from the transport.
type Client struct {
	httpClient *http.Client
}

// NewClient returns a Client that connects to the overlay's session-control
// socket at the given path. The path must be a Unix-domain socket; passing
// an empty string uses DefaultSocketPath.
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
		DisableCompression: true,
	}
	return &Client{httpClient: &http.Client{Transport: transport}}
}

// Kill sends an exact-tuple kill request to the overlay and returns the
// result. The key must be the full, trusted session identity tuple; partial
// or forged tuples will not match a live session. The operation is
// idempotent: killing an already-killed session returns Found=true,
// Killed=false. A session that was never registered or already unregistered
// returns Found=false, Killed=false.
func (c *Client) Kill(ctx context.Context, key sessionkill.Key) (sessionkill.KillResult, error) {
	wire := KillRequest{
		RuntimeCredentialID: key.RuntimeCredentialID,
		NodeID:              key.NodeID,
		BootID:              key.BootID,
		SessionID:           key.SessionID,
	}
	var result KillResult
	if err := c.do(ctx, http.MethodPost, "/v1/sessions/kill", wire, &result); err != nil {
		return sessionkill.KillResult{}, err
	}
	return sessionkill.KillResult{Found: result.Found, Killed: result.Killed}, nil
}

func (c *Client) do(ctx context.Context, method, path string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("sessioncontrol client: encode request: %w", err)
		}
		if len(encoded) > maxRequestBytes {
			return errors.New("sessioncontrol client: request exceeds size limit")
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://unix"+path, body)
	if err != nil {
		return fmt.Errorf("sessioncontrol client: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sessioncontrol client: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))
		return fmt.Errorf("sessioncontrol client: HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxResponseBytes+1)
	encoded, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("sessioncontrol client: read response: %w", err)
	}
	if len(encoded) > maxResponseBytes {
		return errors.New("sessioncontrol client: response exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(responseBody); err != nil {
		return fmt.Errorf("sessioncontrol client: decode response: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("sessioncontrol client: response contains trailing JSON")
	}
	return nil
}

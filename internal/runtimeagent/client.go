package runtimeagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

const maxResponseBytes = 128 << 10

type Client struct {
	httpClient *http.Client
}

func NewClient(socketPath string) *Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socketPath)
		},
		DisableCompression: true,
	}
	return &Client{httpClient: &http.Client{Transport: transport}}
}

func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	var response HealthResponse
	if err := c.do(ctx, http.MethodGet, "/v1/health", nil, &response); err != nil {
		return HealthResponse{}, err
	}
	return response, nil
}

func (c *Client) Inspect(ctx context.Context) (InspectResponse, error) {
	var response InspectResponse
	if err := c.do(ctx, http.MethodGet, "/v1/inspect", nil, &response); err != nil {
		return InspectResponse{}, err
	}
	return response, nil
}

func (c *Client) Validate(ctx context.Context, request ValidateRequest) (ValidateResponse, error) {
	var response ValidateResponse
	if err := c.do(ctx, http.MethodPost, "/v1/validate", request, &response); err != nil {
		return ValidateResponse{}, err
	}
	return response, nil
}

func (c *Client) Apply(ctx context.Context, request ApplyRequest) (ApplyResponse, error) {
	var response ApplyResponse
	if err := c.do(ctx, http.MethodPost, "/v1/apply", request, &response); err != nil {
		return ApplyResponse{}, err
	}
	return response, nil
}

func (c *Client) Rollback(ctx context.Context, request RollbackRequest) (RollbackResponse, error) {
	var response RollbackResponse
	if err := c.do(ctx, http.MethodPost, "/v1/rollback", request, &response); err != nil {
		return RollbackResponse{}, err
	}
	return response, nil
}

func (c *Client) do(ctx context.Context, method, path string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("runtimeagent client: encode request: %w", err)
		}
		if len(encoded) > maxRequestBytes {
			return errors.New("runtimeagent client: request exceeds size limit")
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, "http://unix"+path, body)
	if err != nil {
		return fmt.Errorf("runtimeagent client: build request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("runtimeagent client: request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
		return fmt.Errorf("runtimeagent client: operation failed with HTTP %d", response.StatusCode)
	}

	limited := io.LimitReader(response.Body, maxResponseBytes+1)
	encoded, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("runtimeagent client: read response: %w", err)
	}
	if len(encoded) > maxResponseBytes {
		return errors.New("runtimeagent client: response exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(responseBody); err != nil {
		return fmt.Errorf("runtimeagent client: decode response: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("runtimeagent client: response contains trailing JSON")
	}
	return nil
}

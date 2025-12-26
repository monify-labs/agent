package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/monify-labs/agent/internal/config"
	"github.com/monify-labs/agent/pkg/models"
)

// ErrUnauthorized is returned when authentication fails (401)
var ErrUnauthorized = errors.New("authentication failed: invalid or expired token")

// HTTPSender sends metrics via HTTP/HTTPS
type HTTPSender struct {
	serverURL string
	token     string
	client    *http.Client
}

// NewHTTPSender creates a new HTTP sender
func NewHTTPSender(serverURL, token string) *HTTPSender {
	// Create HTTP client with connection pooling
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HTTPSender{
		serverURL: serverURL,
		token:     token,
		client:    client,
	}
}

// Send sends a single metric payload
func (h *HTTPSender) Send(ctx context.Context, payload *models.MetricPayload) (*models.ServerResponse, error) {
	if payload == nil {
		return nil, nil
	}

	// Marshal to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Compress with gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", h.serverURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("User-Agent", fmt.Sprintf("monify/%s", config.Version))
	req.Header.Set("X-Agent-Version", config.Version)

	// Set authentication if token is configured
	if h.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	}

	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse server response for commands
		var serverResp models.ServerResponse
		if err := json.Unmarshal(respBody, &serverResp); err != nil {
			// If parsing fails, just return success without commands
			return &models.ServerResponse{Status: "success"}, nil
		}
		return &serverResp, nil
	}

	// Handle different error codes
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, fmt.Errorf("bad request: %s", string(respBody))
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limited")
	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}
}

// Close closes the HTTP client
func (h *HTTPSender) Close() error {
	h.client.CloseIdleConnections()
	return nil
}

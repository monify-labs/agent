package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/monify-labs/agent/pkg/config"
	"github.com/monify-labs/agent/pkg/models"
)

// HTTPSender sends metrics via HTTP/HTTPS
type HTTPSender struct {
	config *config.Config
	client *http.Client
}

// NewHTTPSender creates a new HTTP sender
func NewHTTPSender(config *config.Config) *HTTPSender {
	// Configure TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.Server.TLS.InsecureSkipVerify,
	}

	// Create HTTP client with connection pooling
	client := &http.Client{
		Timeout: config.Server.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig:     tlsConfig,
		},
	}

	return &HTTPSender{
		config: config,
		client: client,
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
	req, err := http.NewRequestWithContext(ctx, "POST", h.config.Server.URL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("User-Agent", fmt.Sprintf("monify/%s", h.config.Agent.Version))
	req.Header.Set("X-Agent-Version", h.config.Agent.Version)

	// Set authentication if API key is configured
	if h.config.Server.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.config.Server.Token))
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
		return nil, fmt.Errorf("authentication failed: invalid API key")
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

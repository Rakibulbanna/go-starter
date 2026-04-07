package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"skoolz/config"
)

// HTTPClient represents an HTTP client for external services
type HTTPClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
	timeout time.Duration
}

// NewHTTPClient creates a new HTTP client for external services
func NewHTTPClient(serviceConfig config.ExternalService) *HTTPClient {
	timeout := time.Duration(serviceConfig.Timeout) * time.Second

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: serviceConfig.BaseURL,
		apiKey:  serviceConfig.APIKey,
		timeout: timeout,
	}
}

// Get performs a GET request to the external service
func (c *HTTPClient) Get(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add default headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// Post performs a POST request to the external service
func (c *HTTPClient) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add default headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// Put performs a PUT request to the external service
func (c *HTTPClient) Put(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add default headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// Delete performs a DELETE request to the external service
func (c *HTTPClient) Delete(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add default headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

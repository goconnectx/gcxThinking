// Package systemone provides a Go client for the Local SystemOne (TypeSafe) evaluation API.
package systemone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

const (
	defaultBaseURL   = "http://localhost:11434/v1/systemone"
	defaultUserAgent = "systemone-go"
	defaultTimeout   = 30 * time.Second
)

// Client manages communication with the SystemOne evaluation API.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	userAgent  string
	apiKey     string
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		u, err := url.Parse(baseURL)
		if err == nil {
			c.baseURL = u
		}
	}
}

// WithUserAgent sets a custom user agent string.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new SystemOne API client.
func NewClient(options ...ClientOption) (*Client, error) {
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid default base URL: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    baseURL,
		userAgent:  fmt.Sprintf("%s/%s (%s; %s)", defaultUserAgent, Version, runtime.GOOS, runtime.Version()),
	}

	for _, opt := range options {
		opt(c)
	}

	return c, nil
}

// Evaluate sends a request to the SystemOne evaluation endpoint.
func (c *Client) Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResponse, error) {
	r, err := c.newRequest(ctx, "POST", "", req)
	if err != nil {
		return nil, err
	}

	var resp EvaluateResponse
	if _, err := c.do(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// newRequest creates a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	return req, nil
}

// do executes an HTTP request and decodes the response into v.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if v != nil && len(body) > 0 {
			if err := json.Unmarshal(body, v); err != nil {
				return resp, fmt.Errorf("decoding response: %w", err)
			}
		}
		return resp, nil
	}

	return resp, parseError(resp, body)
}

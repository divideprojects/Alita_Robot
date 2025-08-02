package http

import (
	"context"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// HTTPClient provides a configured HTTP client with proper resource management
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with sensible defaults
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

// Get performs a GET request with proper resource management
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set user agent
	req.Header.Set("User-Agent", "Alita-Robot/2.1.3")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Post performs a POST request with proper resource management
func (c *HTTPClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "Alita-Robot/2.1.3")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetWithTimeout performs a GET request with a specific timeout
func (c *HTTPClient) GetWithTimeout(url string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Get(ctx, url)
}

// SafeReadBody reads the response body and ensures it's properly closed
func SafeReadBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Error("Failed to close response body")
		}
	}()

	return io.ReadAll(resp.Body)
}

// SafeCloseResponse safely closes an HTTP response
func SafeCloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Error("Failed to close HTTP response body")
		}
	}
}

// DownloadFile downloads a file with proper resource management
func (c *HTTPClient) DownloadFile(ctx context.Context, url string, maxSize int64) ([]byte, error) {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer SafeCloseResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    "HTTP request failed",
		}
	}

	// Check content length
	if resp.ContentLength > maxSize {
		return nil, &HTTPError{
			StatusCode: http.StatusRequestEntityTooLarge,
			Message:    "File too large",
		}
	}

	// Use LimitReader to prevent reading too much data
	limitedReader := io.LimitReader(resp.Body, maxSize)
	return io.ReadAll(limitedReader)
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
}

// Close closes the HTTP client and cleans up resources
func (c *HTTPClient) Close() {
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

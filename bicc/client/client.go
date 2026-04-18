package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds the provider configuration
type Config struct {
	Host     string
	Username string
	Password string
	Port     int
}

// Client handles communication with BICC APIs
type Client struct {
	Config     *Config
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient creates a new BICC API client
func NewClient(config *Config) *Client {
	httpClient := &http.Client{
		Timeout: time.Second * 300,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	baseURL := fmt.Sprintf("https://%s:%d", config.Host, config.Port)

	return &Client{
		Config:     config,
		HTTPClient: httpClient,
		BaseURL:    baseURL,
	}
}

// doRequest executes an HTTP request with authentication (context-free, for backwards compat).
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	return c.doRequestWithContext(context.Background(), method, path, body)
}

// doRequestWithContext executes an HTTP request with authentication and a caller-supplied context.
// Using the framework context means the HTTP call is cancelled when Terraform cancels the gRPC
// call, and the gRPC keepalive is satisfied because the provider process is visibly active.
func (c *Client) doRequestWithContext(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	return resp, nil
}

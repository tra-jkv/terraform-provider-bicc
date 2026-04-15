package client

import (
	"bytes"
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

// doRequest executes an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, reqBody)
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

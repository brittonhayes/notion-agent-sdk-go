package notionagents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is the Notion Agents API client.
type Client struct {
	auth          string
	baseURL       string
	notionVersion string
	httpClient    *http.Client
	Agents        *AgentOperations
}

// ClientOptions configures a new Client.
type ClientOptions struct {
	Auth          string       // Required: Notion API token
	BaseURL       string       // Optional: defaults to DefaultBaseURL
	NotionVersion string       // Optional: defaults to DefaultVersion
	HTTPClient    *http.Client // Optional: custom HTTP client
}

// NewClient creates a new Notion Agents client.
func NewClient(opts ClientOptions) *Client {
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}
	if opts.NotionVersion == "" {
		opts.NotionVersion = DefaultVersion
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	c := &Client{
		auth:          opts.Auth,
		baseURL:       strings.TrimRight(opts.BaseURL, "/"),
		notionVersion: opts.NotionVersion,
		httpClient:    opts.HTTPClient,
	}
	c.Agents = &AgentOperations{client: c}
	return c
}

// apiError represents an error response from the Notion API.
type apiError struct {
	Object  string `json:"object"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// doRequest executes an HTTP request with proper headers.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + "/" + strings.TrimLeft(path, "/")

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.auth)
	req.Header.Set("Notion-Version", c.notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// doJSON executes a request and unmarshals the JSON response.
func (c *Client) doJSON(ctx context.Context, method, path string, body, result interface{}) error {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr apiError
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			if apiErr.Code == "object_not_found" {
				if strings.Contains(apiErr.Message, "Could not find agent with ID:") {
					return &AgentNotFoundError{AgentID: extractID(apiErr.Message, "agent")}
				}
				if strings.Contains(apiErr.Message, "Could not find thread with ID:") {
					return &ThreadNotFoundError{ThreadID: extractID(apiErr.Message, "thread")}
				}
			}
			return &NotionAgentsError{Msg: apiErr.Message, Code: apiErr.Code}
		}
		return &NotionAgentsError{
			Msg:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
			Code: "http_error",
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// extractID is a helper to extract an ID from an error message.
func extractID(message, objectType string) string {
	prefix := fmt.Sprintf("Could not find %s with ID: ", objectType)
	if idx := strings.Index(message, prefix); idx >= 0 {
		return strings.TrimSpace(message[idx+len(prefix):])
	}
	return ""
}

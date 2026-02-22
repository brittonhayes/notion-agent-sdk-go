package notionagents

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func mockClient(fn func(req *http.Request) (*http.Response, error)) *Client {
	return NewClient(ClientOptions{
		Auth:       "test-token",
		HTTPClient: &http.Client{Transport: roundTripFunc(fn)},
	})
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.notionVersion != DefaultVersion {
		t.Errorf("notionVersion = %q, want %q", c.notionVersion, DefaultVersion)
	}
	if c.Agents == nil {
		t.Fatal("Agents should not be nil")
	}
}

func TestNewClientCustomOptions(t *testing.T) {
	c := NewClient(ClientOptions{
		Auth:          "tok",
		BaseURL:       "https://custom.api.com/",
		NotionVersion: "2025-01-01",
	})
	if c.baseURL != "https://custom.api.com" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.baseURL)
	}
	if c.notionVersion != "2025-01-01" {
		t.Errorf("notionVersion = %q, want %q", c.notionVersion, "2025-01-01")
	}
}

func TestDoRequestSetsHeaders(t *testing.T) {
	var captured *http.Request
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		captured = req
		return jsonResponse(200, map[string]string{"ok": "true"}), nil
	})

	resp, err := c.doRequest(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if got := captured.Header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
	}
	if got := captured.Header.Get("Notion-Version"); got != DefaultVersion {
		t.Errorf("Notion-Version = %q, want %q", got, DefaultVersion)
	}
}

func TestDoJSONErrorClassification(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		wantErr string
	}{
		{
			name:    "agent not found",
			code:    "object_not_found",
			message: "Could not find agent with ID: abc-123",
			wantErr: "agent not found: abc-123",
		},
		{
			name:    "thread not found",
			code:    "object_not_found",
			message: "Could not find thread with ID: thr-456",
			wantErr: "thread not found: thr-456",
		},
		{
			name:    "generic error",
			code:    "rate_limited",
			message: "Too many requests",
			wantErr: "notion agents error [rate_limited]: Too many requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockClient(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(404, map[string]interface{}{
					"object":  "error",
					"status":  404,
					"code":    tt.code,
					"message": tt.message,
				}), nil
			})

			var result map[string]interface{}
			err := c.doJSON(context.Background(), "GET", "/test", nil, &result)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDoJSONAgentNotFoundType(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(404, map[string]interface{}{
			"object":  "error",
			"status":  404,
			"code":    "object_not_found",
			"message": "Could not find agent with ID: abc-123",
		}), nil
	})

	var result map[string]interface{}
	err := c.doJSON(context.Background(), "GET", "/test", nil, &result)
	if _, ok := err.(*AgentNotFoundError); !ok {
		t.Errorf("expected *AgentNotFoundError, got %T", err)
	}
}

func TestDoJSONThreadNotFoundType(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(404, map[string]interface{}{
			"object":  "error",
			"status":  404,
			"code":    "object_not_found",
			"message": "Could not find thread with ID: thr-456",
		}), nil
	})

	var result map[string]interface{}
	err := c.doJSON(context.Background(), "GET", "/test", nil, &result)
	if _, ok := err.(*ThreadNotFoundError); !ok {
		t.Errorf("expected *ThreadNotFoundError, got %T", err)
	}
}

func TestExtractID(t *testing.T) {
	tests := []struct {
		message    string
		objectType string
		want       string
	}{
		{"Could not find agent with ID: abc-123", "agent", "abc-123"},
		{"Could not find thread with ID: thr-456", "thread", "thr-456"},
		{"Some other error", "agent", ""},
	}

	for _, tt := range tests {
		got := extractID(tt.message, tt.objectType)
		if got != tt.want {
			t.Errorf("extractID(%q, %q) = %q, want %q", tt.message, tt.objectType, got, tt.want)
		}
	}
}

func jsonResponse(statusCode int, body interface{}) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}

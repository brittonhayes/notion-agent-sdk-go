// Package testutil provides mock HTTP clients and response factories
// for testing code that uses the notionagents SDK.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
)

// RoundTripFunc is an adapter to allow ordinary functions as http.RoundTripper.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper.
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// MockHTTPClient returns an *http.Client that uses the given function for all requests.
func MockHTTPClient(fn RoundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

// JSONResponse creates an *http.Response with the given status code and JSON body.
func JSONResponse(statusCode int, body interface{}) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}

// NDJSONResponse creates an *http.Response with NDJSON lines for streaming.
func NDJSONResponse(chunks ...interface{}) *http.Response {
	var lines []string
	for _, chunk := range chunks {
		data, _ := json.Marshal(chunk)
		lines = append(lines, string(data))
	}
	body := strings.Join(lines, "\n") + "\n"
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// ErrorResponse creates an API error response matching Notion's error format.
func ErrorResponse(statusCode int, code, message string) *http.Response {
	return JSONResponse(statusCode, map[string]interface{}{
		"object":  "error",
		"status":  statusCode,
		"code":    code,
		"message": message,
	})
}

// MockAgentListResponse returns a sample AgentListResponse.
func MockAgentListResponse() notionagents.AgentListResponse {
	desc := "A test agent"
	return notionagents.AgentListResponse{
		Object: "list",
		Type:   "agent",
		Results: []notionagents.AgentData{
			{
				Object:      "agent",
				ID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				Name:        "Test Agent",
				Description: &desc,
			},
			{
				Object: "agent",
				ID:     notionagents.PersonalAgentID,
				Name:   "Notion AI",
			},
		},
		HasMore:    false,
		NextCursor: nil,
	}
}

// MockThreadListResponse returns a sample ThreadListResponse.
func MockThreadListResponse() notionagents.ThreadListResponse {
	return notionagents.ThreadListResponse{
		Object: "list",
		Type:   "thread",
		Results: []notionagents.ThreadListItem{
			{
				Object: "thread",
				ID:     "tttttttt-tttt-tttt-tttt-tttttttttttt",
				Title:  "Test Thread",
				Status: notionagents.ThreadStatusCompleted,
				CreatedBy: notionagents.CreatedBy{
					ID:   "uuuuuuuu-uuuu-uuuu-uuuu-uuuuuuuuuuuu",
					Type: "user",
				},
			},
		},
		HasMore:    false,
		NextCursor: nil,
	}
}

// MockChatInvocationResponse returns a sample ChatInvocationResponse.
func MockChatInvocationResponse() notionagents.ChatInvocationResponse {
	return notionagents.ChatInvocationResponse{
		Object:   "chat_invocation",
		AgentID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		ThreadID: "tttttttt-tttt-tttt-tttt-tttttttttttt",
		Status:   "pending",
	}
}

// MockStreamChunks returns sample NDJSON stream chunks (started, message, done).
func MockStreamChunks() []interface{} {
	return []interface{}{
		notionagents.StreamChunk{
			Type:     "started",
			ThreadID: "tttttttt-tttt-tttt-tttt-tttttttttttt",
			AgentID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
		notionagents.StreamChunk{
			Type:    "message",
			ID:      "msg-1",
			Role:    "assistant",
			Content: "Hello",
		},
		notionagents.StreamChunk{
			Type:    "message",
			ID:      "msg-1",
			Role:    "assistant",
			Content: "Hello, world!",
		},
		notionagents.StreamChunk{
			Type: "done",
		},
	}
}

// MockThreadMessageListResponse returns a sample ThreadMessageListResponse.
func MockThreadMessageListResponse() notionagents.ThreadMessageListResponse {
	return notionagents.ThreadMessageListResponse{
		Object: "list",
		Type:   "thread_message",
		Results: []notionagents.ThreadMessageItem{
			{
				Object:  "thread_message",
				ID:      "msg-1",
				Role:    "human",
				Content: "Hello",
				Parent: notionagents.MessageParent{
					Type: "thread",
					ID:   "tttttttt-tttt-tttt-tttt-tttttttttttt",
				},
			},
			{
				Object:  "thread_message",
				ID:      "msg-2",
				Role:    "assistant",
				Content: "Hi there!",
				Parent: notionagents.MessageParent{
					Type: "thread",
					ID:   "tttttttt-tttt-tttt-tttt-tttttttttttt",
				},
			},
		},
		HasMore:    false,
		NextCursor: nil,
	}
}

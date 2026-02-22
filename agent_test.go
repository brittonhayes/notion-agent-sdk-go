package notionagents

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAgentChat(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method = %s, want POST", req.Method)
		}
		if !strings.Contains(req.URL.Path, "/agents/agent-1/chat") {
			t.Errorf("path = %s, want to contain /agents/agent-1/chat", req.URL.Path)
		}

		var body chatRequestBody
		data, _ := io.ReadAll(req.Body)
		json.Unmarshal(data, &body)
		if body.Message != "hello" {
			t.Errorf("body.Message = %q, want %q", body.Message, "hello")
		}

		return jsonResponse(200, ChatInvocationResponse{
			Object:   "chat_invocation",
			AgentID:  "agent-1",
			ThreadID: "thread-1",
			Status:   "pending",
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	resp, err := agent.Chat(context.Background(), ChatParams{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ThreadID != "thread-1" {
		t.Errorf("ThreadID = %q, want %q", resp.ThreadID, "thread-1")
	}
	if resp.Status != "pending" {
		t.Errorf("Status = %q, want %q", resp.Status, "pending")
	}
}

func TestAgentChatValidation(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	agent := c.Agents.Agent("agent-1")

	_, err := agent.Chat(context.Background(), ChatParams{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	ne, ok := err.(*NotionAgentsError)
	if !ok {
		t.Fatalf("expected *NotionAgentsError, got %T", err)
	}
	if ne.Code != "validation_error" {
		t.Errorf("Code = %q, want %q", ne.Code, "validation_error")
	}
}

func TestAgentChatWithAttachments(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		var body chatRequestBody
		data, _ := io.ReadAll(req.Body)
		json.Unmarshal(data, &body)

		if len(body.Attachments) != 1 {
			t.Errorf("attachments len = %d, want 1", len(body.Attachments))
		}
		if body.Attachments[0].FileUpload.ID != "file-1" {
			t.Errorf("attachment file ID = %q, want %q", body.Attachments[0].FileUpload.ID, "file-1")
		}

		return jsonResponse(200, ChatInvocationResponse{
			Object:   "chat_invocation",
			AgentID:  "agent-1",
			ThreadID: "thread-1",
			Status:   "pending",
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	_, err := agent.Chat(context.Background(), ChatParams{
		Message: "check this file",
		Attachments: []ChatAttachmentInput{
			{FileUploadID: "file-1", Name: "test.pdf"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentChatWithThreadID(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		var body chatRequestBody
		data, _ := io.ReadAll(req.Body)
		json.Unmarshal(data, &body)

		if body.ThreadID != "existing-thread" {
			t.Errorf("ThreadID = %q, want %q", body.ThreadID, "existing-thread")
		}

		return jsonResponse(200, ChatInvocationResponse{
			Object:   "chat_invocation",
			AgentID:  "agent-1",
			ThreadID: "existing-thread",
			Status:   "pending",
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	resp, err := agent.Chat(context.Background(), ChatParams{
		Message:  "follow up",
		ThreadID: "existing-thread",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ThreadID != "existing-thread" {
		t.Errorf("ThreadID = %q, want %q", resp.ThreadID, "existing-thread")
	}
}

func TestAgentThread(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	agent := c.Agents.Agent("agent-1")
	thread := agent.Thread("thread-1")

	if thread.ThreadID != "thread-1" {
		t.Errorf("ThreadID = %q, want %q", thread.ThreadID, "thread-1")
	}
	if thread.AgentID != "agent-1" {
		t.Errorf("AgentID = %q, want %q", thread.AgentID, "agent-1")
	}
}

func TestAgentListThreads(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method = %s, want GET", req.Method)
		}
		if !strings.Contains(req.URL.Path, "/agents/agent-1/threads") {
			t.Errorf("path = %s, want to contain /agents/agent-1/threads", req.URL.Path)
		}

		return jsonResponse(200, ThreadListResponse{
			Object: "list",
			Type:   "thread",
			Results: []ThreadListItem{
				{Object: "thread", ID: "t-1", Title: "Thread 1", Status: ThreadStatusCompleted},
			},
			HasMore: false,
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	resp, err := agent.ListThreads(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(resp.Results))
	}
	if resp.Results[0].ID != "t-1" {
		t.Errorf("ID = %q, want %q", resp.Results[0].ID, "t-1")
	}
}

func TestAgentListThreadsWithParams(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		if q.Get("status") != "completed" {
			t.Errorf("status = %q, want %q", q.Get("status"), "completed")
		}
		if q.Get("page_size") != "5" {
			t.Errorf("page_size = %q, want %q", q.Get("page_size"), "5")
		}

		return jsonResponse(200, ThreadListResponse{
			Object:  "list",
			Results: []ThreadListItem{},
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	_, err := agent.ListThreads(context.Background(), &ThreadListParams{
		Status:   ThreadStatusCompleted,
		PageSize: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentGetThread(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadListResponse{
			Object: "list",
			Results: []ThreadListItem{
				{Object: "thread", ID: "t-1", Status: ThreadStatusCompleted},
			},
		}), nil
	})

	agent := c.Agents.Agent("agent-1")
	item, err := agent.GetThread(context.Background(), "t-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "t-1" {
		t.Errorf("ID = %q, want %q", item.ID, "t-1")
	}
}

func TestStreamValidation(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	agent := c.Agents.Agent("agent-1")

	_, err := agent.Stream(context.Background(), ChatStreamParams{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	ne, ok := err.(*NotionAgentsError)
	if !ok {
		t.Fatalf("expected *NotionAgentsError, got %T", err)
	}
	if ne.Code != "validation_error" {
		t.Errorf("Code = %q, want %q", ne.Code, "validation_error")
	}
}

func TestStreamHTTPError(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(`{"error":"server error"}`)),
		}, nil
	})

	agent := c.Agents.Agent("agent-1")
	_, err := agent.Stream(context.Background(), ChatStreamParams{Message: "hello"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	se, ok := err.(*StreamError)
	if !ok {
		t.Fatalf("expected *StreamError, got %T", err)
	}
	if se.Code != "http_error" {
		t.Errorf("Code = %q, want %q", se.Code, "http_error")
	}
}

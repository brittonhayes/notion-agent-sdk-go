package notionagents

import (
	"context"
	"net/http"
	"testing"
)

func TestIterAgentsSinglePage(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, AgentListResponse{
			Object: "list",
			Results: []AgentData{
				{ID: "a-1", Name: "Agent 1"},
				{ID: "a-2", Name: "Agent 2"},
			},
			HasMore: false,
		}), nil
	})

	var agents []AgentData
	for agent, err := range IterAgents(context.Background(), c, nil) {
		if err != nil {
			t.Fatal(err)
		}
		agents = append(agents, agent)
	}

	if len(agents) != 2 {
		t.Fatalf("agents len = %d, want 2", len(agents))
	}
}

func TestIterAgentsMultiPage(t *testing.T) {
	page := 0
	cursor := "cursor-2"
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		page++
		if page == 1 {
			return jsonResponse(200, AgentListResponse{
				Object:     "list",
				Results:    []AgentData{{ID: "a-1"}},
				HasMore:    true,
				NextCursor: &cursor,
			}), nil
		}
		// Verify cursor was passed
		if req.URL.Query().Get("start_cursor") != "cursor-2" {
			t.Errorf("start_cursor = %q, want %q", req.URL.Query().Get("start_cursor"), "cursor-2")
		}
		return jsonResponse(200, AgentListResponse{
			Object:  "list",
			Results: []AgentData{{ID: "a-2"}},
			HasMore: false,
		}), nil
	})

	var agents []AgentData
	for agent, err := range IterAgents(context.Background(), c, nil) {
		if err != nil {
			t.Fatal(err)
		}
		agents = append(agents, agent)
	}

	if len(agents) != 2 {
		t.Fatalf("agents len = %d, want 2", len(agents))
	}
	if agents[0].ID != "a-1" || agents[1].ID != "a-2" {
		t.Errorf("agents = %v, want a-1 then a-2", agents)
	}
}

func TestIterAgentsError(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(500, map[string]interface{}{
			"object":  "error",
			"status":  500,
			"code":    "internal_server_error",
			"message": "something broke",
		}), nil
	})

	var gotErr error
	for _, err := range IterAgents(context.Background(), c, nil) {
		if err != nil {
			gotErr = err
			break
		}
	}

	if gotErr == nil {
		t.Fatal("expected error")
	}
}

func TestIterAgentsEarlyBreak(t *testing.T) {
	cursor := "more"
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, AgentListResponse{
			Object:     "list",
			Results:    []AgentData{{ID: "a-1"}, {ID: "a-2"}, {ID: "a-3"}},
			HasMore:    true,
			NextCursor: &cursor,
		}), nil
	})

	count := 0
	for _, err := range IterAgents(context.Background(), c, nil) {
		if err != nil {
			t.Fatal(err)
		}
		count++
		if count >= 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestCollectAgents(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, AgentListResponse{
			Object:  "list",
			Results: []AgentData{{ID: "a-1"}, {ID: "a-2"}},
			HasMore: false,
		}), nil
	})

	agents, err := CollectAgents(context.Background(), c, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 2 {
		t.Fatalf("agents len = %d, want 2", len(agents))
	}
}

func TestIterThreadsSinglePage(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadListResponse{
			Object: "list",
			Results: []ThreadListItem{
				{ID: "t-1", Title: "Thread 1"},
			},
			HasMore: false,
		}), nil
	})

	agent := c.Agents.Agent("a-1")
	var threads []ThreadListItem
	for thread, err := range IterThreads(context.Background(), agent, nil) {
		if err != nil {
			t.Fatal(err)
		}
		threads = append(threads, thread)
	}

	if len(threads) != 1 {
		t.Fatalf("threads len = %d, want 1", len(threads))
	}
}

func TestCollectThreads(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadListResponse{
			Object:  "list",
			Results: []ThreadListItem{{ID: "t-1"}, {ID: "t-2"}},
			HasMore: false,
		}), nil
	})

	agent := c.Agents.Agent("a-1")
	threads, err := CollectThreads(context.Background(), agent, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 2 {
		t.Fatalf("threads len = %d, want 2", len(threads))
	}
}

func TestIterMessagesSinglePage(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadMessageListResponse{
			Object: "list",
			Results: []ThreadMessageItem{
				{ID: "msg-1", Role: "human", Content: "Hello"},
			},
			HasMore: false,
		}), nil
	})

	thread := &Thread{ThreadID: "t-1", AgentID: "a-1", client: c}
	var messages []ThreadMessageItem
	for msg, err := range IterMessages(context.Background(), thread, nil) {
		if err != nil {
			t.Fatal(err)
		}
		messages = append(messages, msg)
	}

	if len(messages) != 1 {
		t.Fatalf("messages len = %d, want 1", len(messages))
	}
}

func TestCollectMessages(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(200, ThreadMessageListResponse{
			Object:  "list",
			Results: []ThreadMessageItem{{ID: "msg-1"}, {ID: "msg-2"}},
			HasMore: false,
		}), nil
	})

	thread := &Thread{ThreadID: "t-1", AgentID: "a-1", client: c}
	messages, err := CollectMessages(context.Background(), thread, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(messages))
	}
}

func TestIterMessagesMultiPage(t *testing.T) {
	page := 0
	cursor := "page2"
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		page++
		if page == 1 {
			return jsonResponse(200, ThreadMessageListResponse{
				Object:     "list",
				Results:    []ThreadMessageItem{{ID: "msg-1"}},
				HasMore:    true,
				NextCursor: &cursor,
			}), nil
		}
		return jsonResponse(200, ThreadMessageListResponse{
			Object:  "list",
			Results: []ThreadMessageItem{{ID: "msg-2"}},
			HasMore: false,
		}), nil
	})

	thread := &Thread{ThreadID: "t-1", AgentID: "a-1", client: c}
	messages, err := CollectMessages(context.Background(), thread, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(messages))
	}
}

func TestCopyAgentListParamsNil(t *testing.T) {
	p := copyAgentListParams(nil)
	if p == nil {
		t.Fatal("expected non-nil params")
	}
}

func TestCopyAgentListParamsCopy(t *testing.T) {
	original := &AgentListParams{Name: "test", PageSize: 10}
	cp := copyAgentListParams(original)
	cp.Name = "changed"
	if original.Name != "test" {
		t.Error("copy should not modify original")
	}
}

func TestCopyThreadListParamsNil(t *testing.T) {
	p := copyThreadListParams(nil)
	if p == nil {
		t.Fatal("expected non-nil params")
	}
}

func TestCopyMessageListParamsNil(t *testing.T) {
	p := copyMessageListParams(nil)
	if p == nil {
		t.Fatal("expected non-nil params")
	}
}

package notionagents

import (
	"context"
	"net/http"
	"testing"
)

func TestAgentsList(t *testing.T) {
	desc := "A test agent"
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method = %s, want GET", req.Method)
		}

		return jsonResponse(200, AgentListResponse{
			Object: "list",
			Type:   "agent",
			Results: []AgentData{
				{Object: "agent", ID: "a-1", Name: "Agent 1", Description: &desc},
				{Object: "agent", ID: PersonalAgentID, Name: "Notion AI"},
			},
			HasMore: false,
		}), nil
	})

	resp, err := c.Agents.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("results len = %d, want 2", len(resp.Results))
	}
	if resp.Results[0].Name != "Agent 1" {
		t.Errorf("Name = %q, want %q", resp.Results[0].Name, "Agent 1")
	}
}

func TestAgentsListWithParams(t *testing.T) {
	c := mockClient(func(req *http.Request) (*http.Response, error) {
		q := req.URL.Query()
		if q.Get("name") != "search" {
			t.Errorf("name = %q, want %q", q.Get("name"), "search")
		}
		if q.Get("page_size") != "10" {
			t.Errorf("page_size = %q, want %q", q.Get("page_size"), "10")
		}
		if q.Get("start_cursor") != "cursor-abc" {
			t.Errorf("start_cursor = %q, want %q", q.Get("start_cursor"), "cursor-abc")
		}

		return jsonResponse(200, AgentListResponse{
			Object:  "list",
			Results: []AgentData{},
		}), nil
	})

	_, err := c.Agents.List(context.Background(), &AgentListParams{
		Name:        "search",
		PageSize:    10,
		StartCursor: "cursor-abc",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentsAgent(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	agent := c.Agents.Agent("my-agent-id")

	if agent.ID != "my-agent-id" {
		t.Errorf("ID = %q, want %q", agent.ID, "my-agent-id")
	}
}

func TestAgentsPersonal(t *testing.T) {
	c := NewClient(ClientOptions{Auth: "tok"})
	agent := c.Agents.Personal()

	if agent.ID != PersonalAgentID {
		t.Errorf("ID = %q, want %q", agent.ID, PersonalAgentID)
	}
}

package notionagents

import (
	"context"
	"net/url"
	"strconv"
)

// AgentOperations provides operations on agents.
type AgentOperations struct {
	client *Client
}

// List returns a paginated list of agents.
func (a *AgentOperations) List(ctx context.Context, params *AgentListParams) (*AgentListResponse, error) {
	path := "v1/agents"
	if params != nil {
		q := url.Values{}
		if params.Name != "" {
			q.Set("name", params.Name)
		}
		if params.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(params.PageSize))
		}
		if params.StartCursor != "" {
			q.Set("start_cursor", params.StartCursor)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var resp AgentListResponse
	if err := a.client.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Agent returns an Agent handle for the given agent ID.
func (a *AgentOperations) Agent(agentID string) *Agent {
	return &Agent{
		ID:     agentID,
		client: a.client,
	}
}

// Personal returns an Agent handle for the personal agent.
func (a *AgentOperations) Personal() *Agent {
	return a.Agent(PersonalAgentID)
}

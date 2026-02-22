package notionagents

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Thread provides operations on a specific thread.
type Thread struct {
	ThreadID string
	AgentID  string
	client   *Client
}

// Get retrieves this thread's details.
func (t *Thread) Get(ctx context.Context) (*ThreadListItem, error) {
	path := fmt.Sprintf("v1/agents/%s/threads?id=%s", t.AgentID, url.QueryEscape(t.ThreadID))

	var resp ThreadListResponse
	if err := t.client.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, &ThreadNotFoundError{ThreadID: t.ThreadID}
	}

	return &resp.Results[0], nil
}

// Poll polls this thread until completion using exponential backoff.
func (t *Thread) Poll(ctx context.Context, opts *PollThreadOptions) (*ThreadListItem, error) {
	agent := &Agent{ID: t.AgentID, client: t.client}
	return agent.PollThread(ctx, t.ThreadID, opts)
}

// ListMessages returns a paginated list of messages in this thread.
func (t *Thread) ListMessages(ctx context.Context, params *ThreadMessageListParams) (*ThreadMessageListResponse, error) {
	path := fmt.Sprintf("v1/threads/%s/messages", t.ThreadID)
	if params != nil {
		q := url.Values{}
		if params.Verbose != nil {
			q.Set("verbose", strconv.FormatBool(*params.Verbose))
		}
		if params.Role != "" {
			q.Set("role", params.Role)
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

	var resp ThreadMessageListResponse
	if err := t.client.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

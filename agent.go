package notionagents

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

// Agent provides operations on a specific agent.
type Agent struct {
	ID          string
	Name        string
	Instruction *string
	client      *Client
}

// chatRequestBody is the JSON body for chat requests.
type chatRequestBody struct {
	Message     string               `json:"message,omitempty"`
	ThreadID    string               `json:"thread_id,omitempty"`
	Attachments []chatAttachmentBody `json:"attachments,omitempty"`
}

type chatAttachmentBody struct {
	FileUpload fileUploadRef `json:"file_upload"`
	Name       string        `json:"name,omitempty"`
}

type fileUploadRef struct {
	ID string `json:"id"`
}

// Chat starts an async chat with the agent.
func (a *Agent) Chat(ctx context.Context, params ChatParams) (*ChatInvocationResponse, error) {
	if params.Message == "" && len(params.Attachments) == 0 {
		return nil, &NotionAgentsError{
			Msg:  "Either message or attachments is required.",
			Code: "validation_error",
		}
	}

	body := chatRequestBody{
		Message:  params.Message,
		ThreadID: params.ThreadID,
	}
	for _, att := range params.Attachments {
		body.Attachments = append(body.Attachments, chatAttachmentBody{
			FileUpload: fileUploadRef{ID: att.FileUploadID},
			Name:       att.Name,
		})
	}

	var resp ChatInvocationResponse
	path := fmt.Sprintf("v1/agents/%s/chat", a.ID)
	if err := a.client.doJSON(ctx, "POST", path, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Thread returns a Thread handle.
func (a *Agent) Thread(threadID string) *Thread {
	return &Thread{
		ThreadID: threadID,
		AgentID:  a.ID,
		client:   a.client,
	}
}

// GetThread retrieves a specific thread.
func (a *Agent) GetThread(ctx context.Context, threadID string) (*ThreadListItem, error) {
	return a.Thread(threadID).Get(ctx)
}

// PollThread polls a thread until it completes or fails, using exponential backoff.
func (a *Agent) PollThread(ctx context.Context, threadID string, opts *PollThreadOptions) (*ThreadListItem, error) {
	if opts == nil {
		opts = &PollThreadOptions{}
	}
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 60
	}
	baseDelay := opts.BaseDelayMs
	if baseDelay <= 0 {
		baseDelay = 1000
	}
	maxDelay := opts.MaxDelayMs
	if maxDelay <= 0 {
		maxDelay = 10000
	}
	initialDelay := opts.InitialDelayMs
	if initialDelay <= 0 {
		initialDelay = 1000
	}

	// Initial delay before first poll
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(initialDelay) * time.Millisecond):
	}

	thread := a.Thread(threadID)

	for attempt := range maxAttempts {
		item, err := thread.Get(ctx)
		if err != nil {
			if isThreadNotFound(err) {
				if opts.OnThreadNotFound != nil {
					opts.OnThreadNotFound(attempt)
				}
			} else {
				return nil, err
			}
		} else {
			switch item.Status {
			case ThreadStatusCompleted, ThreadStatusFailed:
				return item, nil
			case ThreadStatusPending:
				if opts.OnPending != nil {
					opts.OnPending(*item, attempt)
				}
			}
		}

		// Exponential backoff with jitter
		exponentialDelay := float64(baseDelay) * math.Pow(2, float64(attempt))
		jitter := rand.Float64() * float64(baseDelay)
		delay := math.Min(exponentialDelay+jitter, float64(maxDelay))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(delay) * time.Millisecond):
		}
	}

	return nil, &PollingTimeoutError{Attempts: maxAttempts}
}

// isThreadNotFound checks if an error is a ThreadNotFoundError.
func isThreadNotFound(err error) bool {
	_, ok := err.(*ThreadNotFoundError)
	return ok
}

// ListThreads returns a paginated list of threads for this agent.
func (a *Agent) ListThreads(ctx context.Context, params *ThreadListParams) (*ThreadListResponse, error) {
	path := fmt.Sprintf("v1/agents/%s/threads", a.ID)
	if params != nil {
		q := url.Values{}
		if params.ID != "" {
			q.Set("id", params.ID)
		}
		if params.Title != "" {
			q.Set("title", params.Title)
		}
		if params.Status != "" {
			q.Set("status", string(params.Status))
		}
		if params.CreatedByType != "" {
			q.Set("created_by_type", params.CreatedByType)
		}
		if params.CreatedByID != "" {
			q.Set("created_by_id", params.CreatedByID)
		}
		if params.StartCursor != "" {
			q.Set("start_cursor", params.StartCursor)
		}
		if params.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(params.PageSize))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var resp ThreadListResponse
	if err := a.client.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

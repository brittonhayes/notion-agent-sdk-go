package notionagents

const (
	// PersonalAgentID is the reserved UUID for the personal agent (Notion AI).
	PersonalAgentID = "33333333-3333-3333-3333-333333333333"

	// DefaultBaseURL is the default Notion API base URL.
	DefaultBaseURL = "https://api.notion.com"

	// DefaultVersion is the default Notion API version.
	DefaultVersion = "2025-09-03"
)

// ThreadStatus represents the status of a thread.
type ThreadStatus string

const (
	ThreadStatusPending   ThreadStatus = "pending"
	ThreadStatusCompleted ThreadStatus = "completed"
	ThreadStatusFailed    ThreadStatus = "failed"
)

// AgentVersion contains version information for an agent.
type AgentVersion struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	PublishedAt string `json:"published_at"`
}

// AgentIcon represents an agent's icon, which can be of several types.
type AgentIcon struct {
	Type              string             `json:"type"`
	Emoji             *string            `json:"emoji,omitempty"`
	File              *FileURL           `json:"file,omitempty"`
	External          *ExternalURL       `json:"external,omitempty"`
	CustomEmoji       *CustomEmoji       `json:"custom_emoji,omitempty"`
	CustomAgentAvatar *CustomAgentAvatar `json:"custom_agent_avatar,omitempty"`
}

type FileURL struct {
	URL        string `json:"url"`
	ExpiryTime string `json:"expiry_time,omitempty"`
}

type ExternalURL struct {
	URL string `json:"url"`
}

type CustomEmoji struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type CustomAgentAvatar struct {
	URL string `json:"url"`
}

// AgentData represents an agent returned by the API.
type AgentData struct {
	Object             string        `json:"object"`
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Description        *string       `json:"description"`
	Instruction        *string       `json:"instruction"`
	InstructionsPageID *string       `json:"instructions_page_id"`
	Icon               *AgentIcon    `json:"icon"`
	Version            *AgentVersion `json:"version"`
}

// CreatedBy identifies who created a thread.
type CreatedBy struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ThreadListItem represents a thread in list responses.
type ThreadListItem struct {
	Object       string        `json:"object"`
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	Status       ThreadStatus  `json:"status"`
	CreatedBy    CreatedBy     `json:"created_by"`
	AgentVersion *AgentVersion `json:"agent_version"`
}

// ThreadMessageAttachment represents a file attached to a message.
type ThreadMessageAttachment struct {
	Name        string  `json:"name"`
	ContentType string  `json:"content_type"`
	URL         string  `json:"url"`
	ExpiryTime  *string `json:"expiry_time,omitempty"`
}

// ToolResult represents the result of an agent tool call.
type ToolResult struct {
	ID          string      `json:"id"`
	AgentStepID *string     `json:"agent_step_id"`
	ToolCallID  *string     `json:"tool_call_id"`
	ToolName    string      `json:"tool_name"`
	ToolType    string      `json:"tool_type"`
	State       string      `json:"state"`
	Input       interface{} `json:"input"`
	Output      interface{} `json:"output"`
	Error       *string     `json:"error"`
	StartedAt   int64       `json:"started_at"`
	FinishedAt  *int64      `json:"finished_at"`
	DurationMs  *int64      `json:"duration_ms"`
}

// FollowUp represents a suggested follow-up action.
type FollowUp struct {
	Label   string `json:"label"`
	Message string `json:"message"`
}

// AgentContentPart represents a structured part of agent message content.
type AgentContentPart struct {
	Type       string       `json:"type"`
	Text       string       `json:"text,omitempty"`
	ToolCallID *string      `json:"tool_call_id,omitempty"`
	ToolName   string       `json:"tool_name,omitempty"`
	Input      string       `json:"input,omitempty"`
	Results    []ToolResult `json:"results,omitempty"`
	FollowUps  []FollowUp   `json:"follow_ups,omitempty"`
}

// ThreadMessageItem represents a message within a thread.
type ThreadMessageItem struct {
	Object       string                    `json:"object"`
	ID           string                    `json:"id"`
	Role         string                    `json:"role"`
	Content      string                    `json:"content"`
	Parent       MessageParent             `json:"parent"`
	Attachments  []ThreadMessageAttachment `json:"attachments,omitempty"`
	ContentParts []AgentContentPart        `json:"content_parts,omitempty"`
}

// MessageParent identifies the parent of a message.
type MessageParent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ChatAttachmentInput is used to attach files when sending a chat message.
type ChatAttachmentInput struct {
	FileUploadID string `json:"file_upload_id"`
	Name         string `json:"name,omitempty"`
}

// ChatInvocationResponse is returned when starting an async chat.
type ChatInvocationResponse struct {
	Object   string `json:"object"`
	AgentID  string `json:"agent_id"`
	ThreadID string `json:"thread_id"`
	Status   string `json:"status"`
}

// AgentListParams configures agent listing requests.
type AgentListParams struct {
	Name        string
	PageSize    int
	StartCursor string
}

// AgentListResponse is the paginated response for listing agents.
type AgentListResponse struct {
	Object     string      `json:"object"`
	Type       string      `json:"type"`
	Results    []AgentData `json:"results"`
	HasMore    bool        `json:"has_more"`
	NextCursor *string     `json:"next_cursor"`
}

// ThreadListParams configures thread listing requests.
type ThreadListParams struct {
	ID            string
	Title         string
	Status        ThreadStatus
	CreatedByType string
	CreatedByID   string
	StartCursor   string
	PageSize      int
}

// ThreadListResponse is the paginated response for listing threads.
type ThreadListResponse struct {
	Object     string           `json:"object"`
	Type       string           `json:"type"`
	Results    []ThreadListItem `json:"results"`
	HasMore    bool             `json:"has_more"`
	NextCursor *string          `json:"next_cursor"`
}

// ThreadMessageListParams configures message listing requests.
type ThreadMessageListParams struct {
	Verbose     *bool
	Role        string
	PageSize    int
	StartCursor string
}

// ThreadMessageListResponse is the paginated response for listing messages.
type ThreadMessageListResponse struct {
	Object     string              `json:"object"`
	Type       string              `json:"type"`
	Results    []ThreadMessageItem `json:"results"`
	HasMore    bool                `json:"has_more"`
	NextCursor *string             `json:"next_cursor"`
}

// StreamChunk represents a single chunk from a streaming chat response.
type StreamChunk struct {
	Type         string                    `json:"type"`
	ThreadID     string                    `json:"thread_id,omitempty"`
	AgentID      string                    `json:"agent_id,omitempty"`
	ID           string                    `json:"id,omitempty"`
	Role         string                    `json:"role,omitempty"`
	Content      string                    `json:"content,omitempty"`
	Attachments  []ThreadMessageAttachment `json:"attachments,omitempty"`
	ContentParts []AgentContentPart        `json:"content_parts,omitempty"`
	Code         string                    `json:"code,omitempty"`
	Message      string                    `json:"message,omitempty"`
}

// StreamMessage represents an accumulated message from a stream.
type StreamMessage struct {
	ID           string                    `json:"id"`
	Role         string                    `json:"role"`
	Content      string                    `json:"content"`
	Attachments  []ThreadMessageAttachment `json:"attachments,omitempty"`
	ContentParts []AgentContentPart        `json:"content_parts,omitempty"`
}

// ThreadInfo contains the final result of a completed streaming chat.
type ThreadInfo struct {
	ThreadID string
	AgentID  string
	Messages []StreamMessage
}

// PollThreadOptions configures thread polling behavior.
type PollThreadOptions struct {
	MaxAttempts      int
	BaseDelayMs      int
	MaxDelayMs       int
	InitialDelayMs   int
	OnPending        func(thread ThreadListItem, attempt int)
	OnThreadNotFound func(attempt int)
}

// ChatParams configures a chat request.
type ChatParams struct {
	Message     string
	Attachments []ChatAttachmentInput
	ThreadID    string
}

// ChatStreamParams configures a streaming chat request.
type ChatStreamParams struct {
	Message     string
	Attachments []ChatAttachmentInput
	ThreadID    string
	Verbose     *bool
	OnMessage   func(message StreamMessage)
}

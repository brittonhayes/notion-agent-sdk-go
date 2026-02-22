package notionagents

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StreamReader reads streaming chat responses using an iterator pattern.
type StreamReader struct {
	resp      *http.Response
	scanner   *bufio.Scanner
	threadID  string
	agentID   string
	messages  map[string]*StreamMessage
	msgOrder  []string
	done      bool
	onMessage func(StreamMessage)
}

// Next returns the next chunk from the stream.
// Returns io.EOF when the stream is complete.
func (r *StreamReader) Next() (StreamChunk, error) {
	if r.done {
		return StreamChunk{}, io.EOF
	}

	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == "" {
			continue
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return StreamChunk{}, &StreamError{
				Msg:  fmt.Sprintf("failed to parse stream chunk: %s", err),
				Code: "invalid_stream_response",
			}
		}

		switch chunk.Type {
		case "started":
			r.threadID = chunk.ThreadID
			r.agentID = chunk.AgentID

		case "message":
			if _, exists := r.messages[chunk.ID]; !exists {
				r.msgOrder = append(r.msgOrder, chunk.ID)
			}
			r.messages[chunk.ID] = &StreamMessage{
				ID:           chunk.ID,
				Role:         chunk.Role,
				Content:      chunk.Content,
				Attachments:  chunk.Attachments,
				ContentParts: chunk.ContentParts,
			}
			if r.onMessage != nil {
				r.onMessage(*r.messages[chunk.ID])
			}

		case "done":
			r.done = true
			return chunk, nil

		case "error":
			return StreamChunk{}, &StreamError{
				Msg:  chunk.Message,
				Code: chunk.Code,
			}
		}

		return chunk, nil
	}

	if err := r.scanner.Err(); err != nil {
		return StreamChunk{}, &StreamError{
			Msg:  fmt.Sprintf("scanner error: %s", err),
			Code: "stream_read_error",
		}
	}

	return StreamChunk{}, io.EOF
}

// Close closes the underlying response body.
func (r *StreamReader) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// ThreadInfo returns the accumulated thread info after the stream completes.
func (r *StreamReader) ThreadInfo() *ThreadInfo {
	if r.threadID == "" {
		return nil
	}

	msgs := make([]StreamMessage, 0, len(r.msgOrder))
	for _, id := range r.msgOrder {
		if msg, ok := r.messages[id]; ok {
			msgs = append(msgs, *msg)
		}
	}

	return &ThreadInfo{
		ThreadID: r.threadID,
		AgentID:  r.agentID,
		Messages: msgs,
	}
}

// Stream opens a streaming chat connection and returns a StreamReader.
func (a *Agent) Stream(ctx context.Context, params ChatStreamParams) (*StreamReader, error) {
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

	path := fmt.Sprintf("v1/agents/%s/chatStream", a.ID)

	// Add verbose query param (default true)
	verbose := true
	if params.Verbose != nil {
		verbose = *params.Verbose
	}
	if verbose {
		path += "?verbose=true"
	}

	resp, err := a.client.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &StreamError{
			Msg:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
			Code: "http_error",
		}
	}

	if resp.Body == nil {
		return nil, &StreamError{
			Msg:  "response body is nil",
			Code: "missing_response_body",
		}
	}

	return &StreamReader{
		resp:      resp,
		scanner:   bufio.NewScanner(resp.Body),
		messages:  make(map[string]*StreamMessage),
		onMessage: params.OnMessage,
	}, nil
}

// ChatStream opens a streaming chat and returns channels for chunks, thread info, and errors.
func (a *Agent) ChatStream(ctx context.Context, params ChatStreamParams) (<-chan StreamChunk, <-chan *ThreadInfo, <-chan error) {
	chunks := make(chan StreamChunk)
	info := make(chan *ThreadInfo, 1)
	errc := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(info)
		defer close(errc)

		reader, err := a.Stream(ctx, params)
		if err != nil {
			errc <- err
			return
		}
		defer reader.Close()

		for {
			chunk, err := reader.Next()
			if err == io.EOF {
				if ti := reader.ThreadInfo(); ti != nil {
					info <- ti
				}
				return
			}
			if err != nil {
				errc <- err
				return
			}
			select {
			case chunks <- chunk:
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			}
		}
	}()

	return chunks, info, errc
}

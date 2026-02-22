package notionagents

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func makeStreamReader(chunks ...StreamChunk) *StreamReader {
	var lines []string
	for _, c := range chunks {
		data, _ := json.Marshal(c)
		lines = append(lines, string(data))
	}
	body := strings.Join(lines, "\n") + "\n"

	return &StreamReader{
		scanner:  bufio.NewScanner(strings.NewReader(body)),
		messages: make(map[string]*StreamMessage),
	}
}

func TestStreamReaderBasicFlow(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "Hello"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "Hello, world!"},
		StreamChunk{Type: "done"},
	)

	// started
	chunk, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Type != "started" {
		t.Errorf("type = %q, want %q", chunk.Type, "started")
	}

	// first message
	chunk, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Type != "message" {
		t.Errorf("type = %q, want %q", chunk.Type, "message")
	}
	if chunk.Content != "Hello" {
		t.Errorf("content = %q, want %q", chunk.Content, "Hello")
	}

	// updated message
	chunk, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Content != "Hello, world!" {
		t.Errorf("content = %q, want %q", chunk.Content, "Hello, world!")
	}

	// done
	chunk, err = r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Type != "done" {
		t.Errorf("type = %q, want %q", chunk.Type, "done")
	}

	// EOF after done
	_, err = r.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestStreamReaderThreadInfo(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "Hello"},
		StreamChunk{Type: "message", ID: "msg-2", Role: "assistant", Content: "Follow up"},
		StreamChunk{Type: "done"},
	)

	// Consume all chunks
	for {
		_, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	info := r.ThreadInfo()
	if info == nil {
		t.Fatal("ThreadInfo should not be nil")
	}
	if info.ThreadID != "t-1" {
		t.Errorf("ThreadID = %q, want %q", info.ThreadID, "t-1")
	}
	if info.AgentID != "a-1" {
		t.Errorf("AgentID = %q, want %q", info.AgentID, "a-1")
	}
	if len(info.Messages) != 2 {
		t.Fatalf("messages len = %d, want 2", len(info.Messages))
	}
	if info.Messages[0].Content != "Hello" {
		t.Errorf("msg[0].Content = %q, want %q", info.Messages[0].Content, "Hello")
	}
	if info.Messages[1].Content != "Follow up" {
		t.Errorf("msg[1].Content = %q, want %q", info.Messages[1].Content, "Follow up")
	}
}

func TestStreamReaderMessageAccumulation(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "H"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "He"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "Hello"},
		StreamChunk{Type: "done"},
	)

	for {
		_, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	info := r.ThreadInfo()
	if len(info.Messages) != 1 {
		t.Fatalf("messages len = %d, want 1 (accumulated)", len(info.Messages))
	}
	if info.Messages[0].Content != "Hello" {
		t.Errorf("final content = %q, want %q", info.Messages[0].Content, "Hello")
	}
}

func TestStreamReaderErrorChunk(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "error", Code: "agent_error", Message: "something broke"},
	)

	// started
	_, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}

	// error chunk
	_, err = r.Next()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	se, ok := err.(*StreamError)
	if !ok {
		t.Fatalf("expected *StreamError, got %T", err)
	}
	if se.Code != "agent_error" {
		t.Errorf("Code = %q, want %q", se.Code, "agent_error")
	}
	if se.Msg != "something broke" {
		t.Errorf("Msg = %q, want %q", se.Msg, "something broke")
	}
}

func TestStreamReaderInvalidJSON(t *testing.T) {
	r := &StreamReader{
		scanner:  bufio.NewScanner(strings.NewReader("not valid json\n")),
		messages: make(map[string]*StreamMessage),
	}

	_, err := r.Next()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	se, ok := err.(*StreamError)
	if !ok {
		t.Fatalf("expected *StreamError, got %T", err)
	}
	if se.Code != "invalid_stream_response" {
		t.Errorf("Code = %q, want %q", se.Code, "invalid_stream_response")
	}
}

func TestStreamReaderEmptyStream(t *testing.T) {
	r := &StreamReader{
		scanner:  bufio.NewScanner(strings.NewReader("")),
		messages: make(map[string]*StreamMessage),
	}

	_, err := r.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestStreamReaderSkipsBlankLines(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
	)
	// Inject blank lines by recreating scanner with blanks
	r.scanner = bufio.NewScanner(strings.NewReader(
		"\n\n" + mustJSON(StreamChunk{Type: "started", ThreadID: "t-1"}) + "\n\n",
	))

	chunk, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if chunk.Type != "started" {
		t.Errorf("type = %q, want %q", chunk.Type, "started")
	}
}

func TestStreamReaderOnMessageCallback(t *testing.T) {
	var callbackMessages []StreamMessage
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "message", ID: "msg-1", Role: "assistant", Content: "Hi"},
		StreamChunk{Type: "done"},
	)
	r.onMessage = func(msg StreamMessage) {
		callbackMessages = append(callbackMessages, msg)
	}

	for {
		_, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	if len(callbackMessages) != 1 {
		t.Fatalf("callback called %d times, want 1", len(callbackMessages))
	}
	if callbackMessages[0].Content != "Hi" {
		t.Errorf("callback content = %q, want %q", callbackMessages[0].Content, "Hi")
	}
}

func TestStreamReaderThreadInfoBeforeStarted(t *testing.T) {
	r := &StreamReader{
		scanner:  bufio.NewScanner(strings.NewReader("")),
		messages: make(map[string]*StreamMessage),
	}

	info := r.ThreadInfo()
	if info != nil {
		t.Errorf("expected nil ThreadInfo before stream started, got %+v", info)
	}
}

func TestStreamReaderClose(t *testing.T) {
	r := &StreamReader{}
	if err := r.Close(); err != nil {
		t.Errorf("Close on nil resp should not error, got %v", err)
	}
}

func TestStreamReaderMessageOrder(t *testing.T) {
	r := makeStreamReader(
		StreamChunk{Type: "started", ThreadID: "t-1", AgentID: "a-1"},
		StreamChunk{Type: "message", ID: "msg-a", Role: "assistant", Content: "First"},
		StreamChunk{Type: "message", ID: "msg-b", Role: "assistant", Content: "Second"},
		StreamChunk{Type: "message", ID: "msg-c", Role: "assistant", Content: "Third"},
		StreamChunk{Type: "done"},
	)

	for {
		_, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	info := r.ThreadInfo()
	if len(info.Messages) != 3 {
		t.Fatalf("messages len = %d, want 3", len(info.Messages))
	}

	expected := []string{"First", "Second", "Third"}
	for i, want := range expected {
		if info.Messages[i].Content != want {
			t.Errorf("msg[%d].Content = %q, want %q", i, info.Messages[i].Content, want)
		}
	}
}

func mustJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

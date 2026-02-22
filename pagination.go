package notionagents

import (
	"context"
	"iter"
)

// IterAgents returns an iterator over all agents, automatically handling pagination.
func IterAgents(ctx context.Context, client *Client, params *AgentListParams) iter.Seq2[AgentData, error] {
	return func(yield func(AgentData, error) bool) {
		p := copyAgentListParams(params)
		for {
			resp, err := client.Agents.List(ctx, p)
			if err != nil {
				yield(AgentData{}, err)
				return
			}
			for _, agent := range resp.Results {
				if !yield(agent, nil) {
					return
				}
			}
			if !resp.HasMore || resp.NextCursor == nil {
				return
			}
			p.StartCursor = *resp.NextCursor
		}
	}
}

// CollectAgents collects all agents into a slice.
func CollectAgents(ctx context.Context, client *Client, params *AgentListParams) ([]AgentData, error) {
	var agents []AgentData
	for agent, err := range IterAgents(ctx, client, params) {
		if err != nil {
			return agents, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// IterThreads returns an iterator over all threads for an agent.
func IterThreads(ctx context.Context, agent *Agent, params *ThreadListParams) iter.Seq2[ThreadListItem, error] {
	return func(yield func(ThreadListItem, error) bool) {
		p := copyThreadListParams(params)
		for {
			resp, err := agent.ListThreads(ctx, p)
			if err != nil {
				yield(ThreadListItem{}, err)
				return
			}
			for _, thread := range resp.Results {
				if !yield(thread, nil) {
					return
				}
			}
			if !resp.HasMore || resp.NextCursor == nil {
				return
			}
			p.StartCursor = *resp.NextCursor
		}
	}
}

// CollectThreads collects all threads into a slice.
func CollectThreads(ctx context.Context, agent *Agent, params *ThreadListParams) ([]ThreadListItem, error) {
	var threads []ThreadListItem
	for thread, err := range IterThreads(ctx, agent, params) {
		if err != nil {
			return threads, err
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

// IterMessages returns an iterator over all messages in a thread.
func IterMessages(ctx context.Context, thread *Thread, params *ThreadMessageListParams) iter.Seq2[ThreadMessageItem, error] {
	return func(yield func(ThreadMessageItem, error) bool) {
		p := copyMessageListParams(params)
		for {
			resp, err := thread.ListMessages(ctx, p)
			if err != nil {
				yield(ThreadMessageItem{}, err)
				return
			}
			for _, msg := range resp.Results {
				if !yield(msg, nil) {
					return
				}
			}
			if !resp.HasMore || resp.NextCursor == nil {
				return
			}
			p.StartCursor = *resp.NextCursor
		}
	}
}

// CollectMessages collects all messages into a slice.
func CollectMessages(ctx context.Context, thread *Thread, params *ThreadMessageListParams) ([]ThreadMessageItem, error) {
	var messages []ThreadMessageItem
	for msg, err := range IterMessages(ctx, thread, params) {
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func copyAgentListParams(p *AgentListParams) *AgentListParams {
	if p == nil {
		return &AgentListParams{}
	}
	cp := *p
	return &cp
}

func copyThreadListParams(p *ThreadListParams) *ThreadListParams {
	if p == nil {
		return &ThreadListParams{}
	}
	cp := *p
	return &cp
}

func copyMessageListParams(p *ThreadMessageListParams) *ThreadMessageListParams {
	if p == nil {
		return &ThreadMessageListParams{}
	}
	cp := *p
	return &cp
}

package main

import (
	"strings"

	notionagents "github.com/brittonhayes/notion-agent-sdk-go"
	"github.com/charmbracelet/glamour"
)

type markdownRenderer struct {
	renderer *glamour.TermRenderer
	width    int
}

func newMarkdownRenderer(width int) *markdownRenderer {
	if width < 20 {
		width = 80
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4),
	)

	return &markdownRenderer{
		renderer: r,
		width:    width,
	}
}

func (m *markdownRenderer) render(content string) string {
	if m.renderer == nil {
		return content
	}

	cleaned := notionagents.StripLangTags(content)
	rendered, err := m.renderer.Render(cleaned)
	if err != nil {
		return content
	}
	return strings.TrimSpace(rendered)
}

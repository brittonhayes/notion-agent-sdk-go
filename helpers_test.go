package notionagents

import "testing"

func TestIsPersonalAgent(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{PersonalAgentID, true},
		{"33333333-3333-3333-3333-333333333333", true},
		{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsPersonalAgent(tt.id)
		if got != tt.want {
			t.Errorf("IsPersonalAgent(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestStripLangTags(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`<lang en>Hello</lang>`, `Hello</lang>`},
		{`<lang fr="true">Bonjour</lang>`, `Bonjour</lang>`},
		{`No tags here`, `No tags here`},
		{``, ``},
		{`<lang en>First</lang> and <lang fr>Second</lang>`, `First</lang> and Second</lang>`},
		{`<lang  multiple="a" attrs="b">text`, `text`},
	}

	for _, tt := range tests {
		got := StripLangTags(tt.input)
		if got != tt.want {
			t.Errorf("StripLangTags(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

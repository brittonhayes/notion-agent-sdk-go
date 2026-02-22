package notionagents

import "regexp"

var langTagRegexp = regexp.MustCompile(`<lang\s[^>]*>`)

// IsPersonalAgent returns true if the given agent ID is the personal agent.
func IsPersonalAgent(agentID string) bool {
	return agentID == PersonalAgentID
}

// StripLangTags removes <lang ...> XML tags from text.
func StripLangTags(text string) string {
	return langTagRegexp.ReplaceAllString(text, "")
}

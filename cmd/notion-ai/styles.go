package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Notion-inspired color palette
var (
	// Accent colors
	notionBlue   = lipgloss.Color("#2EAADC")
	notionRed    = lipgloss.Color("#EB5757")
	notionGreen  = lipgloss.Color("#6BC950")
	notionYellow = lipgloss.Color("#DFAB01")

	// Light theme
	lightBg      = lipgloss.Color("#FFFFFF")
	lightFg      = lipgloss.Color("#37352F")
	lightMuted   = lipgloss.Color("#9B9A97")
	lightBorder  = lipgloss.Color("#E9E9E7")
	lightSurface = lipgloss.Color("#F7F6F3")

	// Dark theme
	darkBg      = lipgloss.Color("#191919")
	darkFg      = lipgloss.Color("#E3E2E0")
	darkMuted   = lipgloss.Color("#6B6B6B")
	darkBorder  = lipgloss.Color("#373737")
	darkSurface = lipgloss.Color("#252525")
)

type themeColors struct {
	bg      lipgloss.Color
	fg      lipgloss.Color
	muted   lipgloss.Color
	border  lipgloss.Color
	surface lipgloss.Color
	accent  lipgloss.Color
}

func lightTheme() themeColors {
	return themeColors{
		bg:      lightBg,
		fg:      lightFg,
		muted:   lightMuted,
		border:  lightBorder,
		surface: lightSurface,
		accent:  notionBlue,
	}
}

func darkTheme() themeColors {
	return themeColors{
		bg:      darkBg,
		fg:      darkFg,
		muted:   darkMuted,
		border:  darkBorder,
		surface: darkSurface,
		accent:  notionBlue,
	}
}

type styles struct {
	app              lipgloss.Style
	header           lipgloss.Style
	headerTitle      lipgloss.Style
	headerInfo       lipgloss.Style
	viewport         lipgloss.Style
	humanLabel       lipgloss.Style
	agentLabel       lipgloss.Style
	humanMsg         lipgloss.Style
	agentMsg         lipgloss.Style
	inputArea        lipgloss.Style
	inputAreaFocused lipgloss.Style
	statusBar        lipgloss.Style
	statusKey        lipgloss.Style
	statusVal        lipgloss.Style
	errorStyle       lipgloss.Style
	spinnerStyle     lipgloss.Style
	divider          lipgloss.Style
	listTitle        lipgloss.Style
	listItem         lipgloss.Style
	listItemSel      lipgloss.Style
	listItemDesc     lipgloss.Style
	streamingText    lipgloss.Style
}

func newStyles(theme themeColors) styles {
	return styles{
		app: lipgloss.NewStyle(),

		header: lipgloss.NewStyle().
			Foreground(theme.muted).
			Padding(0, 2, 0, 2),

		headerTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.fg),

		headerInfo: lipgloss.NewStyle().
			Foreground(theme.muted),

		viewport: lipgloss.NewStyle().
			Padding(0, 1),

		humanLabel: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.fg),

		agentLabel: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.accent),

		humanMsg: lipgloss.NewStyle().
			Foreground(theme.fg),

		agentMsg: lipgloss.NewStyle().
			Foreground(theme.fg),

		inputArea: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.border).
			Padding(0, 1),

		inputAreaFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1),

		statusBar: lipgloss.NewStyle().
			Foreground(theme.muted).
			Padding(0, 2),

		statusKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true),

		statusVal: lipgloss.NewStyle().
			Foreground(theme.muted),

		errorStyle: lipgloss.NewStyle().
			Foreground(notionRed).
			Bold(true),

		spinnerStyle: lipgloss.NewStyle().
			Foreground(theme.accent),

		divider: lipgloss.NewStyle().
			Foreground(theme.border),

		listTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.accent).
			Padding(0, 0, 1, 0),

		listItem: lipgloss.NewStyle().
			Foreground(theme.fg).
			Padding(0, 2),

		listItemSel: lipgloss.NewStyle().
			Foreground(theme.accent).
			Bold(true).
			Padding(0, 2),

		listItemDesc: lipgloss.NewStyle().
			Foreground(theme.muted),

		streamingText: lipgloss.NewStyle().
			Foreground(theme.fg),
	}
}

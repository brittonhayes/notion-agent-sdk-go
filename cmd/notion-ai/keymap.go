package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit        key.Binding
	Send        key.Binding
	ToggleFocus key.Binding
	Cancel      key.Binding
	Select      key.Binding
	Back        key.Binding
	Filter      key.Binding
	NewInList   key.Binding
	ScrollUp    key.Binding
	ScrollDown  key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Send: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	),
	ToggleFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle focus"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel/back"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	NewInList: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new thread"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("up/k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("down/j", "scroll down"),
	),
}

package ui

import "strings"

// KeyBinding represents a single key binding with its description
type KeyBinding struct {
	Key         string
	Description string
}

// ViewKeyBindings holds the key bindings for a specific view
type ViewKeyBindings struct {
	AllowedKeys []string     // Keys that are allowed in this view (excluding global keys)
	StatusBar   []KeyBinding // Keys to show in the status bar
}

// Global key bindings that work in all views
var GlobalKeys = []KeyBinding{
	{"?", "help"},
	{"q", "quit / go back (2x in feed view)"},
	{"esc", "go back (no-op in feed view)"},
	{"ctrl+c", "go back / quit (2x in feed view)"},
	{"j, down", "move down"},
	{"k, up", "move up"},
	{"enter", "select / open"},
	{"ctrl+d", "page down"},
	{"ctrl+u", "page up"},
}

// View-specific key bindings
var FeedListViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"r", "R", "l", "t", "c", "U", "u", "i", "/", "ctrl+f"},
	StatusBar: []KeyBinding{
		{"/", "search"},
		{"c", "config"},
		{"r/R", "reload"},
	},
}

var ItemListViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"r", "R", "A", "/", "ctrl+f", "h", "l", "left", "right", "0", "$"},
	StatusBar: []KeyBinding{
		{"/", "search"},
		{"r/R", "reload"},
		{"h/l, ←/→, 0/$", "navigate title"},
	},
}

var ArticleViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "n", "N", "o", "r"},
	StatusBar: []KeyBinding{
		{"n/N", "next/prev"},
	}, // No custom status bar for article view
}

var SettingsViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"?"},
	StatusBar: []KeyBinding{
		{"?", "settings help"},
	},
}

var TasksViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"c", "d", "l"},
	StatusBar: []KeyBinding{
		{Key: "A", Description: "clear failed"},
	},
}

var FeedInfoViewKeys = ViewKeyBindings{
	AllowedKeys: []string{},
	StatusBar:   []KeyBinding{},
}

var LogViewKeys = ViewKeyBindings{
	AllowedKeys: []string{"c"},
	StatusBar: []KeyBinding{
		{Key: "A", Description: "clear all"},
	},
}

var URLsViewKeys = ViewKeyBindings{
	AllowedKeys: []string{},
	StatusBar:   []KeyBinding{},
}

var HelpViewKeys = ViewKeyBindings{
	AllowedKeys: []string{},
	StatusBar:   []KeyBinding{},
}

// GetViewKeys returns the key bindings for a given view state
func GetViewKeys(state ViewState) ViewKeyBindings {
	switch state {
	case FeedListView:
		return FeedListViewKeys
	case ItemListView:
		return ItemListViewKeys
	case ArticleView:
		return ArticleViewKeys
	case FeedInfoView:
		return FeedInfoViewKeys
	case SettingsView:
		return SettingsViewKeys
	case TasksView:
		return TasksViewKeys
	case LogView:
		return LogViewKeys
	case URLsView:
		return URLsViewKeys
	case HelpView:
		return HelpViewKeys
	default:
		return ViewKeyBindings{}
	}
}

// FormatStatusBar creates a formatted status bar string from key bindings
func FormatStatusBar(bindings []KeyBinding) string {
	if len(bindings) == 0 {
		return ""
	}

	parts := make([]string, len(bindings))
	for i, binding := range bindings {
		parts[i] = binding.Key + ": " + binding.Description
	}
	return strings.Join(parts, " | ")
}

package tui

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		category string
		action   string
		target   string
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
		{
			name:     "category alias c",
			input:    "c list",
			category: "containers",
			action:   "list",
		},
		{
			name:     "category alias db",
			input:    "db browse",
			category: "databases",
			action:   "browse",
		},
		{
			name:     "category alias p",
			input:    "p kill",
			category: "processes",
			action:   "kill",
		},
		{
			name:     "full category name",
			input:    "containers list",
			category: "containers",
			action:   "list",
		},
		{
			name:     "category with target",
			input:    "containers stop nginx",
			category: "containers",
			action:   "stop",
			target:   "nginx",
		},
		{
			name:     "category with multi-word target",
			input:    "containers stop my app",
			category: "containers",
			action:   "stop",
			target:   "my app",
		},
		{
			name:     "short-form verb stop",
			input:    "stop nginx",
			category: "containers",
			action:   "stop",
			target:   "nginx",
		},
		{
			name:     "short-form verb kill",
			input:    "kill node",
			category: "processes",
			action:   "kill",
			target:   "node",
		},
		{
			name:     "short-form verb browse",
			input:    "browse mydb",
			category: "databases",
			action:   "browse",
			target:   "mydb",
		},
		{
			name:     "short-form verb no target",
			input:    "stop",
			category: "containers",
			action:   "stop",
		},
		{
			name:   "unknown command",
			input:  "foobar",
			action: "foobar",
		},
		{
			name:   "unknown command with args",
			input:  "foobar baz qux",
			action: "foobar",
			target: "baz qux",
		},
		{
			name:     "case insensitive category",
			input:    "Containers List",
			category: "containers",
			action:   "list",
		},
		{
			name:     "case insensitive verb",
			input:    "STOP nginx",
			category: "containers",
			action:   "stop",
			target:   "nginx",
		},
		{
			name:     "category only no action",
			input:    "containers",
			category: "containers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommand(tt.input)
			if got.Category != tt.category {
				t.Errorf("Category = %q, want %q", got.Category, tt.category)
			}
			if got.Action != tt.action {
				t.Errorf("Action = %q, want %q", got.Action, tt.action)
			}
			if got.Target != tt.target {
				t.Errorf("Target = %q, want %q", got.Target, tt.target)
			}
			if got.Raw != tt.input {
				t.Errorf("Raw = %q, want %q", got.Raw, tt.input)
			}
		})
	}
}

func TestFilterPrefix(t *testing.T) {
	candidates := []string{"containers", "processes", "databases", "help", "quit"}

	tests := []struct {
		name   string
		prefix string
		want   int
	}{
		{"empty prefix matches all", "", 5},
		{"matches con", "con", 1},
		{"matches p", "p", 1},
		{"matches d", "d", 1},
		{"no match", "xyz", 0},
		{"case insensitive", "CON", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPrefix(candidates, tt.prefix)
			if len(got) != tt.want {
				t.Errorf("filterPrefix(%q) len = %d, want %d", tt.prefix, len(got), tt.want)
			}
		})
	}
}

package processor

import (
	"strings"
	"testing"
)

func TestRemoveFillers(t *testing.T) {
	tests := []struct {
		input string
		check func(string) bool
		desc  string
	}{
		{
			"uh I want to create a function",
			func(s string) bool { return !strings.Contains(strings.ToLower(s), "uh") },
			"removes uh",
		},
		{
			"so basically um we need a handler",
			func(s string) bool {
				return !strings.Contains(strings.ToLower(s), "basically") &&
					!strings.Contains(strings.ToLower(s), " um ")
			},
			"removes basically and um",
		},
		{"no fillers here", func(s string) bool { return s == "no fillers here" }, "no change"},
	}

	for _, tt := range tests {
		got := RemoveFillers(tt.input)
		if !tt.check(got) {
			t.Errorf("RemoveFillers(%q) = %q — failed check: %s", tt.input, got, tt.desc)
		}
	}
}

func TestCollapseSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello   world", "hello world"},
		{"  spaces  everywhere  ", " spaces everywhere "},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		got := CollapseSpaces(tt.input)
		if got != tt.expected {
			t.Errorf("CollapseSpaces(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"", ""},
		{"123abc", "123abc"},
	}

	for _, tt := range tests {
		got := CapitalizeFirst(tt.input)
		if got != tt.expected {
			t.Errorf("CapitalizeFirst(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestEnsurePeriod(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world."},
		{"hello world.", "hello world."},
		{"is this a question?", "is this a question?"},
		{"wow!", "wow!"},
		{"", ""},
	}

	for _, tt := range tests {
		got := EnsurePeriod(tt.input)
		if got != tt.expected {
			t.Errorf("EnsurePeriod(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestPipeline(t *testing.T) {
	p := DefaultPipeline()
	p.Add(EnsurePeriod)

	input := "  uh I want to   create um a function  "
	got := p.Process(input)

	if got == "" {
		t.Fatal("pipeline returned empty string")
	}

	if got[0] < 'A' || got[0] > 'Z' {
		t.Errorf("expected capitalized, got %q", got)
	}

	if got[len(got)-1] != '.' {
		t.Errorf("expected period at end, got %q", got)
	}

	if strings.Contains(got, "  ") {
		t.Errorf("should not contain double spaces: %q", got)
	}
}

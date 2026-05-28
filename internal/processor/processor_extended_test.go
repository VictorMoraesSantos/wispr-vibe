package processor

import (
	"strings"
	"testing"
)

func TestRemoveFillersPTBR(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		checkFn func(string) bool
		desc    string
	}{
		{
			"removes tipo assim",
			"tipo assim precisamos de algo",
			func(s string) bool { return !strings.Contains(s, "tipo assim") },
			"should remove 'tipo assim'",
		},
		{
			"removes um (filler pattern)",
			"uh um eu quero algo",
			func(s string) bool { return !strings.Contains(s, "uh") && !strings.Contains(s, " um ") },
			"should remove uh and um fillers",
		},
		{
			"keeps non-filler text",
			"precisamos criar algo novo",
			func(s string) bool { return s == "precisamos criar algo novo" },
			"text without fillers should stay unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveFillers(tt.input)
			if !tt.checkFn(got) {
				t.Errorf("RemoveFillers(%q) = %q — failed: %s", tt.input, got, tt.desc)
			}
		})
	}
}

func TestRemoveFillersEN(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"removes you know", "you know we need to fix this", " we need to fix this"},
		{"removes basically", "so basically the server is down", "so  the server is down"},
		{"removes multiple fillers", "uh I want to um create a function", " I want to  create a function"},
		{"removes so yeah", "so yeah that is the plan", " that is the plan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveFillers(tt.input)
			if got != tt.expect {
				t.Errorf("RemoveFillers(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestTrimText(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"  hello  ", "hello"},
		{"\t\n text \n\t", "text"},
		{"no trim", "no trim"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TrimText(tt.input)
			if got != tt.expect {
				t.Errorf("TrimText(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestCapitalizeFirstUnicode(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"ação", "Ação"},
		{"über", "Über"},
		{"café", "Café"},
		{"日本語", "日本語"},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CapitalizeFirst(tt.input)
			if got != tt.expect {
				t.Errorf("CapitalizeFirst(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestEnsurePeriodVariousPunctuation(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"hello", "hello."},
		{"hello.", "hello."},
		{"hello!", "hello!"},
		{"hello?", "hello?"},
		{"hello:", "hello:"},
		{"hello;", "hello;"},
		{"", ""},
		{"Sentence with comma,", "Sentence with comma,."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EnsurePeriod(tt.input)
			if got != tt.expect {
				t.Errorf("EnsurePeriod(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestPipelineOrder(t *testing.T) {
	var order []int
	f1 := func(s string) string { order = append(order, 1); return s + "A" }
	f2 := func(s string) string { order = append(order, 2); return s + "B" }
	f3 := func(s string) string { order = append(order, 3); return s + "C" }

	p := NewPipeline(f1, f2, f3)
	result := p.Process("X")

	if result != "XABC" {
		t.Errorf("Process result = %q, want %q", result, "XABC")
	}
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("order = %v, want [1,2,3]", order)
	}
}

func TestPipelineAdd(t *testing.T) {
	p := NewPipeline()
	p.Add(func(s string) string { return s + "!" })
	result := p.Process("hello")
	if result != "hello!" {
		t.Errorf("result = %q, want %q", result, "hello!")
	}
}

func TestDefaultPipelineEndToEnd(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			"full cleanup",
			"  uh I want to   create um a function  ",
			"I want to create a function",
		},
		{
			"already clean",
			"Create a function that returns a string",
			"Create a function that returns a string",
		},
		{
			"empty input",
			"",
			"",
		},
		{
			"only fillers",
			"uh um basically",
			"",
		},
	}

	p := DefaultPipeline()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Process(tt.input)
			if got != tt.expect {
				t.Errorf("Process(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestDefaultPipelineWithPeriod(t *testing.T) {
	p := DefaultPipeline()
	p.Add(EnsurePeriod)

	got := p.Process("  hello world  ")
	if got != "Hello world." {
		t.Errorf("result = %q, want %q", got, "Hello world.")
	}
}

package agents

import "testing"

func TestSanitizeCodexOutput(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "code block with next",
			content:  "Here is code:\n```go\npackage main\n\nfunc main() {}\n```\n\nNEXT: done",
			expected: "package main\n\nfunc main() {}",
		},
		{
			name:     "plain code with next line",
			content:  "package main\n\nNEXT: done",
			expected: "package main",
		},
		{
			name:     "plain code only",
			content:  "package main\n\nfunc main() {}",
			expected: "package main\n\nfunc main() {}",
		},
		{
			name:     "empty",
			content:  " \n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeCodexOutput(tt.content)
			if got != tt.expected {
				t.Errorf("sanitizeCodexOutput(%q) = %q, want %q", tt.content, got, tt.expected)
			}
		})
	}
}

func TestParseCodexFileBlocks(t *testing.T) {
	content := "FILE: a/b.go\npackage main\n\nfunc main() {}\nEND_FILE\nFILE: README.md\n# Hello\nEND_FILE\n"
	blocks := parseCodexFileBlocks(content)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].path != "a/b.go" {
		t.Errorf("first path = %q, want %q", blocks[0].path, "a/b.go")
	}
	if blocks[1].path != "README.md" {
		t.Errorf("second path = %q, want %q", blocks[1].path, "README.md")
	}
	if blocks[0].content == "" || blocks[1].content == "" {
		t.Errorf("expected non-empty content for both blocks")
	}
}

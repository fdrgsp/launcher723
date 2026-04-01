package main

import (
	"os"
	"path/filepath"
	"testing"
)

var selectRunnerTests = []struct {
	name     string
	filename string
	content  string
	expected string
}{
	{"ipynb uses juv", "notebook.ipynb", "", "uvx juv run"},
	{"py with marimo dep uses marimo", "nb.py", "# dependencies = [\n#   \"marimo\",\n# ]", "uvx marimo edit --sandbox"},
	{"py without marimo uses uv run", "script.py", "# dependencies = [\n#   \"numpy\",\n# ]", "uv run"},
	{"py with empty content uses uv run", "script.py", "", "uv run"},
	{"py with marimo in comment string uses marimo", "nb.py", "# requires = [\"marimo>=0.1\"]", "uvx marimo edit --sandbox"},
}

func TestSelectRunner(t *testing.T) {
	for _, tc := range selectRunnerTests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, tc.filename)
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			got := selectRunner(path)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

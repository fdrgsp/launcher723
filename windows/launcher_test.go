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
	{"py with marimo dep edit mode", "nb.py", "# /// script\n# dependencies = [\n#   \"marimo\",\n# ]\n#\n# [pyrunner]\n# marimo-mode = \"edit\"\n# ///\n", "uvx marimo edit --sandbox"},
	{"py with marimo dep run mode", "nb.py", "# /// script\n# dependencies = [\n#   \"marimo\",\n# ]\n#\n# [pyrunner]\n# marimo-mode = \"run\"\n# ///\n", "uvx marimo run --sandbox"},
	{"py without marimo uses uv run", "script.py", "# dependencies = [\n#   \"numpy\",\n# ]", "uv run"},
	{"py with empty content uses uv run", "script.py", "", "uv run"},
	{"py with marimo version spec edit mode", "nb.py", "# /// script\n# dependencies = [\n#   \"marimo>=0.1\",\n# ]\n#\n# [pyrunner]\n# marimo-mode = \"edit\"\n# ///\n", "uvx marimo edit --sandbox"},
	{"py with single-quoted marimo edit mode", "nb.py", "# /// script\n# dependencies = [\n#   'marimo',\n# ]\n#\n# [pyrunner]\n# marimo-mode = \"edit\"\n# ///\n", "uvx marimo edit --sandbox"},
	{"py with unrelated marimo mention uses uv run", "script.py", "# this is not marimo_extra related", "uv run"},
	{"py with marimo dep no pyrunner section defaults to edit", "nb.py", "# /// script\n# dependencies = [\n#   \"marimo\",\n# ]\n# ///\n", "uvx marimo edit --sandbox"},
	{"py with marimo dep run mode no spaces", "nb.py", "# /// script\n# dependencies = [\n#   \"marimo\",\n# ]\n#\n#[pyrunner]\n#marimo-mode = \"run\"\n# ///\n", "uvx marimo run --sandbox"},
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

var marimoModeTests = []struct {
	name     string
	content  string
	expected string
}{
	{"no script block", "# dependencies = [\n#   \"marimo\",\n# ]", ""},
	{"run mode", "# /// script\n# [pyrunner]\n# marimo-mode = \"run\"\n# ///\n", "run"},
	{"edit mode", "# /// script\n# [pyrunner]\n# marimo-mode = \"edit\"\n# ///\n", "edit"},
	{"single-quoted run mode", "# /// script\n# [pyrunner]\n# marimo-mode = 'run'\n# ///\n", "run"},
	{"no pyrunner section", "# /// script\n# dependencies = [\n#   \"marimo\",\n# ]\n# ///\n", ""},
	{"section without marimo-mode", "# /// script\n# [pyrunner]\n# other_key = \"value\"\n# ///\n", ""},
	{"marimo-mode after other keys", "# /// script\n# [pyrunner]\n# other = \"x\"\n# marimo-mode = \"run\"\n# ///\n", "run"},
}

func TestMarimoMode(t *testing.T) {
	for _, tc := range marimoModeTests {
		t.Run(tc.name, func(t *testing.T) {
			got := marimoMode(tc.content)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}


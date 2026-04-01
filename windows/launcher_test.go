package main

import (
	"os"
	"path/filepath"
	"testing"
)

var findNotebooksTests = []struct {
	name     string
	files    []string
	expected []string
}{
	{"empty dir", []string{}, nil},
	{"single ipynb", []string{"notebook.ipynb"}, []string{"notebook.ipynb"}},
	{"single py", []string{"script.py"}, []string{"script.py"}},
	{"ipynb and py", []string{"a.ipynb", "b.py"}, []string{"a.ipynb", "b.py"}},
	{"ignores non-notebook files", []string{"nb.ipynb", "readme.txt", "data.csv"}, []string{"nb.ipynb"}},
	{"multiple notebooks", []string{"a.ipynb", "b.ipynb", "c.py"}, []string{"a.ipynb", "b.ipynb", "c.py"}},
}

func TestFindNotebooks(t *testing.T) {
	for _, tc := range findNotebooksTests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
			got := findNotebooks(dir)
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
			for i, want := range tc.expected {
				if got[i] != want {
					t.Errorf("[%d] expected %q, got %q", i, want, got[i])
				}
			}
		})
	}
}

var pickerBehaviorTests = []struct {
	name         string
	files        []string
	expectPicker bool
	expectFile   string
}{
	{"no notebooks shows picker", []string{}, true, ""},
	{"one notebook skips picker", []string{"nb.ipynb"}, false, "nb.ipynb"},
	{"two notebooks shows picker", []string{"a.ipynb", "b.py"}, true, ""},
	{"only non-notebook files shows picker", []string{"readme.txt"}, true, ""},
	{"three notebooks shows picker", []string{"a.ipynb", "b.ipynb", "c.py"}, true, ""},
}

func TestPickerBehavior(t *testing.T) {
	for _, tc := range pickerBehaviorTests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0644); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
			matches := findNotebooks(dir)
			needsPicker := len(matches) != 1
			if needsPicker != tc.expectPicker {
				t.Errorf("expected needsPicker=%v, got %v (matches=%v)", tc.expectPicker, needsPicker, matches)
			}
			if !needsPicker && tc.expectFile != "" && matches[0] != tc.expectFile {
				t.Errorf("expected file %q, got %q", tc.expectFile, matches[0])
			}
		})
	}
}

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

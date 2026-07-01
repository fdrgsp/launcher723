package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ipynbWithHeader builds a minimal .ipynb JSON string with headerLines as the
// source of the hidden PEP 723 metadata cell — mirrors what `juv add` produces.
func ipynbWithHeader(headerLines []string) string {
	nb := map[string]any{
		"cells": []map[string]any{
			{
				"cell_type":       "code",
				"execution_count": nil,
				"id":              "meta",
				"metadata":        map[string]any{"jupyter": map[string]any{"source_hidden": true}},
				"outputs":         []any{},
				"source":          headerLines,
			},
		},
		"metadata":       map[string]any{},
		"nbformat":       4,
		"nbformat_minor": 5,
	}
	b, err := json.MarshalIndent(nb, "", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

var selectRunnerTests = []struct {
	name     string
	filename string
	content  string
	expected string
}{
	{"ipynb uses juv", "notebook.ipynb", "", "uvx juv run"},
	{"ipynb with juv-mode exec", "notebook.ipynb", ipynbWithHeader([]string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = \"exec\"\n", "# ///"}), "uvx juv exec"},
	{"ipynb with juv-mode run explicit", "notebook.ipynb", ipynbWithHeader([]string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = \"run\"\n", "# ///"}), "uvx juv run"},
	{"ipynb with no pyrunner section defaults to run", "notebook.ipynb", ipynbWithHeader([]string{"# /// script\n", "# dependencies = [\n", "#   \"numpy\",\n", "# ]\n", "# ///"}), "uvx juv run"},
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

var juvModeTests = []struct {
	name     string
	header   []string
	expected string
}{
	{"no script block", []string{"print('hi')\n"}, ""},
	{"run mode", []string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = \"run\"\n", "# ///"}, "run"},
	{"exec mode", []string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = \"exec\"\n", "# ///"}, "exec"},
	{"single-quoted exec mode", []string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = 'exec'\n", "# ///"}, "exec"},
	{"no pyrunner section", []string{"# /// script\n", "# dependencies = [\n", "#   \"numpy\",\n", "# ]\n", "# ///"}, ""},
	{"section without juv-mode", []string{"# /// script\n", "# [pyrunner]\n", "# other_key = \"value\"\n", "# ///"}, ""},
	{"juv-mode after other keys", []string{"# /// script\n", "# [pyrunner]\n", "# other = \"x\"\n", "# juv-mode = \"exec\"\n", "# ///"}, "exec"},
	{"trailing inline comment with another quoted mode value", []string{"# /// script\n", "# [pyrunner]\n", "# juv-mode = \"exec\"  # or \"run\" (default)\n", "# ///"}, "exec"},
	{"malformed json defaults to empty", []string{}, ""},
}

func TestJuvMode(t *testing.T) {
	for _, tc := range juvModeTests {
		t.Run(tc.name, func(t *testing.T) {
			got := juvMode(ipynbWithHeader(tc.header))
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}

	t.Run("invalid json returns empty", func(t *testing.T) {
		if got := juvMode("not json"); got != "" {
			t.Errorf("expected empty string for invalid JSON, got %q", got)
		}
	})
}

func TestBuildBatchScript(t *testing.T) {
	script := buildBatchScript(`C:\Users\me\nbs`, "demo.ipynb", "uvx juv run")

	// The bootstrap must skip install when uv is present, pin the install dir
	// to a known location, and add exactly that dir to PATH so the freshly
	// installed uv is found without reopening the terminal.
	wantLines := []string{
		`where uv >nul 2>&1 && goto run`,
		`if not defined UV_INSTALL_DIR set "UV_INSTALL_DIR=%USERPROFILE%\.local\bin"`,
		`set "PATH=%UV_INSTALL_DIR%;%PATH%"`,
		`:run`,
		`cd /d "C:\Users\me\nbs"`,
		`uvx juv run "demo.ipynb"`,
		`pause`,
	}
	for _, want := range wantLines {
		if !strings.Contains(script, "\n"+want+"\n") {
			t.Errorf("script missing line %q\n--- script ---\n%s", want, script)
		}
	}

	// Batch commands and especially the :run label must sit at column 0; a
	// leading tab/space (e.g. from a raw-string reindent) can break goto.
	for _, line := range strings.Split(script, "\n") {
		if line != strings.TrimLeft(line, " \t") {
			t.Errorf("batch line is indented (would corrupt the .bat): %q", line)
		}
	}
}

func TestBuildBatchScriptEscapesPercent(t *testing.T) {
	// '%' in the dir/name must be doubled so cmd.exe treats it literally
	// instead of as a %VAR% reference.
	script := buildBatchScript(`C:\50%done`, "ab%cd.py", "uv run")
	if !strings.Contains(script, `cd /d "C:\50%%done"`) {
		t.Errorf("directory '%%' not doubled in:\n%s", script)
	}
	if !strings.Contains(script, `uv run "ab%%cd.py"`) {
		t.Errorf("filename '%%' not doubled in:\n%s", script)
	}
}

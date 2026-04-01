// Windows .exe launcher — opens a file picker for .ipynb/.py files, then runs
// with uvx juv run (Jupyter), uvx marimo edit --sandbox (marimo), or uv run (plain .py).

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// findNotebooks returns the names of .ipynb and .py files found in dir.
func findNotebooks(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var matches []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ipynb") || strings.HasSuffix(e.Name(), ".py") {
			matches = append(matches, e.Name())
		}
	}
	return matches
}

// selectRunner returns the run command for the given notebook file path.
func selectRunner(notebookPath string) string {
	if strings.HasSuffix(notebookPath, ".ipynb") {
		return "uvx juv run"
	}
	content, _ := os.ReadFile(notebookPath)
	if strings.Contains(string(content), `"marimo`) {
		return "uvx marimo edit --sandbox"
	}
	return "uv run"
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exe)

	notebook := ""
	notebookDir := exeDir

	matches := findNotebooks(exeDir)
	if len(matches) == 1 {
		notebook = matches[0]
	}

	if notebook == "" {
		// Show file picker for .ipynb and .py files
		out, err := exec.Command("powershell", "-NoProfile", "-Command", `
			Add-Type -AssemblyName System.Windows.Forms
			$dlg = New-Object System.Windows.Forms.OpenFileDialog
			$dlg.Title = "Select a notebook (.ipynb or .py)"
			$dlg.Filter = "Notebooks (*.ipynb;*.py)|*.ipynb;*.py|Jupyter Notebooks (*.ipynb)|*.ipynb|Python Scripts (*.py)|*.py"
			$dlg.InitialDirectory = [Environment]::GetFolderPath('UserProfile')
			if ($dlg.ShowDialog() -eq 'OK') { $dlg.FileName } else { "" }
		`).Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error showing file dialog: %v\n", err)
			os.Exit(1)
		}

		selected := strings.TrimSpace(string(out))
		if selected == "" {
			os.Exit(0)
		}

		notebookDir = filepath.Dir(selected)
		notebook = filepath.Base(selected)
	}

	runCmd := selectRunner(filepath.Join(notebookDir, notebook))

	// Bootstrap uv if needed, then run
	tmpDir := os.TempDir()
	script := fmt.Sprintf(`@echo off
where uv >nul 2>&1 || (
    echo Installing uv...
    powershell -ExecutionPolicy Bypass -c "irm https://astral.sh/uv/install.ps1 | iex"
)
cd /d "%s"
echo Launching %s ...
%s "%s"
pause
`, notebookDir, notebook, runCmd, notebook)

	batPath := filepath.Join(tmpDir, "notebook-launcher-run.bat")
	os.WriteFile(batPath, []byte(script), 0644)

	cmd := exec.Command("cmd", "/c", batPath)
	cmd.Dir = notebookDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	os.Remove(batPath)
}

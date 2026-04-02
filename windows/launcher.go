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

// selectRunner returns the run command for the given notebook file path.
func selectRunner(notebookPath string) string {
	if strings.HasSuffix(notebookPath, ".ipynb") {
		return "uvx juv run"
	}
	content, err := os.ReadFile(notebookPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v\n", notebookPath, err)
		return "uv run"
	}
	if isMarimo(string(content)) {
		return "uvx marimo edit --sandbox"
	}
	return "uv run"
}

// isMarimo reports whether file content declares a marimo dependency.
// It matches PEP 723 / TOML patterns like "marimo", 'marimo', "marimo>=1.0".
func isMarimo(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		// Match common dependency declaration patterns:
		//   "marimo"  "marimo>=1.0"  'marimo'  'marimo>=1.0'
		if strings.Contains(trimmed, `"marimo"`) ||
			strings.Contains(trimmed, `"marimo>`) ||
			strings.Contains(trimmed, `"marimo<`) ||
			strings.Contains(trimmed, `"marimo=`) ||
			strings.Contains(trimmed, `"marimo~`) ||
			strings.Contains(trimmed, `"marimo!`) ||
			strings.Contains(trimmed, `'marimo'`) ||
			strings.Contains(trimmed, `'marimo>`) ||
			strings.Contains(trimmed, `'marimo<`) ||
			strings.Contains(trimmed, `'marimo=`) ||
			strings.Contains(trimmed, `'marimo~`) ||
			strings.Contains(trimmed, `'marimo!`) {
			return true
		}
	}
	return false
}

func main() {
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

	notebookDir := filepath.Dir(selected)
	notebook := filepath.Base(selected)
	runCmd := selectRunner(selected)

	// Sanitize values for safe batch-file interpolation: double any '%' so
	// cmd.exe doesn't treat them as variable references, and quote paths.
	safeDirArg := strings.ReplaceAll(notebookDir, "%", "%%")
	safeNameArg := strings.ReplaceAll(notebook, "%", "%%")

	// Bootstrap uv if needed, then run
	script := fmt.Sprintf(`@echo off
powershell -ExecutionPolicy Bypass -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser -Force" >nul 2>&1
where uv >nul 2>&1 || (
    echo Installing uv...
    powershell -ExecutionPolicy Bypass -c "irm https://astral.sh/uv/install.ps1 | iex"
)
cd /d "%s"
echo Launching %s ...
%s "%s"
pause
`, safeDirArg, safeNameArg, runCmd, safeNameArg)

	// Use a unique temp file to avoid races when launched multiple times.
	batFile, err := os.CreateTemp("", "pyrunner-*.bat")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	batPath := batFile.Name()
	batFile.WriteString(script)
	batFile.Close()

	cmd := exec.Command("cmd", "/c", batPath)
	cmd.Dir = notebookDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()

	os.Remove(batPath)
}

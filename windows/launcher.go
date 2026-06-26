// Windows .exe launcher — opens a file picker for .ipynb/.py files, then runs
// with uvx juv run (Jupyter), uvx marimo run/edit --sandbox (marimo), or uv run (plain .py).

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var marimoModeRe = regexp.MustCompile(`^#\s*marimo-mode\s*=\s*["']([a-z]+)["']`)

// marimoMode reads the [pyrunner] section inside the # /// script block and
// returns the marimo-mode value ("run", "edit", or "" if not set).
func marimoMode(content string) string {
	inBlock := false
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		if !inBlock {
			if line == "# /// script" {
				inBlock = true
			}
		} else if !inSection {
			if line == "# ///" {
				break
			}
			if strings.TrimSpace(strings.TrimPrefix(line, "#")) == "[pyrunner]" {
				inSection = true
			}
		} else {
			if line == "# ///" {
				break
			}
			// Another TOML section starts — stop looking.
			if rest := strings.TrimSpace(strings.TrimPrefix(line, "#")); strings.HasPrefix(rest, "[") {
				break
			}
			if m := marimoModeRe.FindStringSubmatch(line); m != nil {
				return m[1]
			}
		}
	}
	return ""
}

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
	s := string(content)
	if isMarimo(s) {
		if marimoMode(s) == "run" {
			return "uvx marimo run --sandbox"
		}
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

// buildBatchScript returns the cmd.exe batch script that bootstraps uv (if
// missing) and runs the notebook with runCmd (e.g. "uvx juv run").
//
// If uv is missing we pin UV_INSTALL_DIR (unless the user already set it) so
// the installer lands uv in a known location, then add exactly that dir to the
// current session's PATH — the installer only updates PATH for *future* shells,
// so without this the freshly-installed uv wouldn't be found until the terminal
// is reopened.
//
// notebookDir/notebook are interpolated with any '%' doubled so cmd.exe treats
// them literally rather than as variable references.
func buildBatchScript(notebookDir, notebook, runCmd string) string {
	safeDir := strings.ReplaceAll(notebookDir, "%", "%%")
	safeName := strings.ReplaceAll(notebook, "%", "%%")
	return fmt.Sprintf(`@echo off
powershell -ExecutionPolicy Bypass -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser -Force" >nul 2>&1
where uv >nul 2>&1 && goto run
echo Installing uv...
if not defined UV_INSTALL_DIR set "UV_INSTALL_DIR=%%USERPROFILE%%\.local\bin"
powershell -ExecutionPolicy Bypass -c "irm https://astral.sh/uv/install.ps1 | iex"
set "PATH=%%UV_INSTALL_DIR%%;%%PATH%%"
:run
cd /d "%s"
echo Launching %s ...
%s "%s"
pause
`, safeDir, safeName, runCmd, safeName)
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

	runCmd := selectRunner(selected)
	notebookDir := filepath.Dir(selected)
	notebook := filepath.Base(selected)

	script := buildBatchScript(notebookDir, notebook, runCmd)

	// Use a unique temp file to avoid races when launched multiple times.
	batFile, err := os.CreateTemp("", "pyrunner-*.bat")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	batPath := batFile.Name()
	batFile.WriteString(script)
	batFile.Close()
	defer os.Remove(batPath)

	cmd := exec.Command("cmd", "/c", batPath)
	cmd.Dir = notebookDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

#!/bin/bash
# macOS .app launcher — opens a file picker for .ipynb/.py files, then runs
# with uvx juv run (Jupyter), uvx marimo edit --sandbox (marimo), or uv run (plain .py).

# Outputs the run command for the given notebook file path.
select_runner() {
  local notebook="$1"
  case "$notebook" in
    *.ipynb) echo "uvx juv run" ;;
    *.py)
      # Match PEP 723 dependency patterns: "marimo", "marimo>=1", 'marimo', etc.
      # Anchored to quote + "marimo" + (quote or version specifier) to avoid
      # false positives on unrelated strings containing "marimo".
      if grep -qE "[\"']marimo[\"'><=~!]" "$notebook"; then
        echo "uvx marimo edit --sandbox"
      else
        echo "uv run"
      fi
      ;;
  esac
}

_main() {
  local NOTEBOOK
  NOTEBOOK=$(osascript -e 'try
    set theFile to choose file with prompt "Select a notebook (.ipynb or .py):" of type {"public.item"} default location (path to home folder)
    return POSIX path of theFile
  on error
    return ""
  end try' 2>/dev/null)

  if [ -z "$NOTEBOOK" ]; then
    exit 0
  fi

  # Verify it's a supported file type
  case "$NOTEBOOK" in
    *.ipynb|*.py) ;;
    *)
      osascript -e 'display alert "Error" message "Please select a .ipynb or .py file."'
      exit 1
      ;;
  esac

  local NOTEBOOK_DIR NOTEBOOK_NAME RUN_CMD
  NOTEBOOK_DIR="$(dirname "$NOTEBOOK")"
  NOTEBOOK_NAME="$(basename "$NOTEBOOK")"
  RUN_CMD="$(select_runner "$NOTEBOOK")"

  # Build a temp runner script.  Values are injected via printf '%q' (shell-
  # escaped) so that crafted filenames cannot break out of the script.
  local RUNNER
  RUNNER=$(mktemp /tmp/pyrunner.XXXXXX)
  {
    echo '#!/bin/bash'
    printf 'NB_DIR=%q\n'  "$NOTEBOOK_DIR"
    printf 'NB_NAME=%q\n' "$NOTEBOOK_NAME"
    printf 'NB_CMD=%q\n'  "$RUN_CMD"
    printf 'NB_SELF=%q\n' "$RUNNER"
    cat << 'BODY'
export PATH="$HOME/.local/bin:$PATH"
if ! command -v uv >/dev/null 2>&1; then
  echo "Installing uv..."
  curl -LsSf https://astral.sh/uv/install.sh | sh
  export PATH="$HOME/.local/bin:$PATH"
fi
cd -- "$NB_DIR"
echo "Launching $NB_NAME ..."
eval "$NB_CMD" "$NB_NAME"
rm -f "$NB_SELF"
BODY
  } > "$RUNNER"
  chmod +x "$RUNNER"

  open -a Terminal "$RUNNER"
}

# Allow sourcing this file for testing without running the launcher.
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then _main; fi

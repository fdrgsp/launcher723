#!/bin/bash
# macOS .app launcher — opens a file picker for .ipynb/.py files, then runs
# with uvx juv run/exec (Jupyter), uvx marimo run/edit --sandbox (marimo), or uv run (plain .py).

# Extracts marimo_mode from the # /// pyrunner block in a .py file.
# Outputs "run", "edit", or "" if not specified.
marimo_mode() {
  local file="$1" in_block=0 in_section=0
  local re=$'^#[[:space:]]*marimo-mode[[:space:]]*=[[:space:]]*["\']([a-z]+)["\']'
  while IFS= read -r line; do
    if [[ $in_block -eq 0 ]]; then
      [[ "$line" == "# /// script" ]] && in_block=1
    elif [[ $in_section -eq 0 ]]; then
      [[ "$line" == "# ///" ]] && break
      [[ "${line#\#}" =~ ^[[:space:]]*\[pyrunner\] ]] && in_section=1
    else
      [[ "$line" == "# ///" ]] && break
      [[ "$line" =~ ^#[[:space:]]*\[ ]] && break
      if [[ "$line" =~ $re ]]; then
        echo "${BASH_REMATCH[1]}"
        return
      fi
    fi
  done < "$file"
}

# Extracts juv_mode from the # /// script block inside an .ipynb file's hidden
# metadata cell. Unlike .py files, that block is embedded as a JSON array of
# quoted, backslash-escaped source lines rather than raw text, so each line is
# unwrapped/unescaped before running the same [pyrunner] scan used for marimo.
# Outputs "run", "exec", or "" if not specified.
juv_mode() {
  local file="$1" in_block=0 in_section=0 line content
  local unwrap=$'^[[:space:]]*"(.*)"[,]?[[:space:]]*$'
  local re=$'^#[[:space:]]*juv-mode[[:space:]]*=[[:space:]]*["\']([a-z]+)["\']'
  while IFS= read -r line; do
    [[ "$line" =~ $unwrap ]] || continue
    content="${BASH_REMATCH[1]}"
    content="${content%\\n}"
    content="${content//\\\"/\"}"
    if [[ $in_block -eq 0 ]]; then
      [[ "$content" == "# /// script" ]] && in_block=1
    elif [[ $in_section -eq 0 ]]; then
      [[ "$content" == "# ///" ]] && break
      [[ "${content#\#}" =~ ^[[:space:]]*\[pyrunner\] ]] && in_section=1
    else
      [[ "$content" == "# ///" ]] && break
      [[ "$content" =~ ^#[[:space:]]*\[ ]] && break
      if [[ "$content" =~ $re ]]; then
        echo "${BASH_REMATCH[1]}"
        return
      fi
    fi
  done < "$file"
}

# Outputs the run command for the given notebook file path.
select_runner() {
  local notebook="$1"
  case "$notebook" in
    *.ipynb)
      if [[ "$(juv_mode "$notebook")" == "exec" ]]; then
        echo "uvx juv exec"
      else
        echo "uvx juv run"
      fi
      ;;
    *.py)
      # Match PEP 723 dependency patterns: "marimo", "marimo>=1", 'marimo', etc.
      # Anchored to quote + "marimo" + (quote or version specifier) to avoid
      # false positives on unrelated strings containing "marimo".
      if grep -qE "[\"']marimo[\"'><=~!]" "$notebook"; then
        if [[ "$(marimo_mode "$notebook")" == "run" ]]; then
          echo "uvx marimo run --sandbox"
        else
          echo "uvx marimo edit --sandbox"
        fi
      else
        echo "uv run"
      fi
      ;;
  esac
}

_main() {
  local NOTEBOOK

  if [ -n "$1" ] && [ -f "$1" ]; then
    NOTEBOOK="$1"
  else
    NOTEBOOK=$(osascript -e 'try
      set theFile to choose file with prompt "Select a notebook (.ipynb or .py):" of type {"public.item"} default location (path to home folder)
      return POSIX path of theFile
    on error
      return ""
    end try' 2>/dev/null)
  fi

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
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then _main "$@"; fi

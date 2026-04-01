"""Tests for windows/launcher.go — delegates to `go test` when Go is available."""

import os
import shutil
import subprocess

import pytest

WINDOWS_DIR = os.path.join(os.path.dirname(__file__), "..", "windows")


@pytest.fixture(scope="session")
def go_available():
    return shutil.which("go") is not None


def _go_test(go_available, run_filter):
    if not go_available:
        pytest.skip("go not installed")
    result = subprocess.run(
        ["go", "test", "-v", "-run", run_filter, "-count=1", "./..."],
        cwd=WINDOWS_DIR,
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, (
        f"Go test {run_filter!r} failed:\n{result.stdout}\n{result.stderr}"
    )
    assert "=== RUN" in result.stdout, (
        f"No tests matched {run_filter!r} — did the test name change?\n{result.stdout}"
    )


# ── findNotebooks ──────────────────────────────────────────────────────────────
# Go spaces → underscores in -run filter

@pytest.mark.parametrize("case", [
    "empty_dir",
    "single_ipynb",
    "single_py",
    "ipynb_and_py",
    "ignores_non-notebook_files",
    "multiple_notebooks",
])
def test_find_notebooks(go_available, case):
    _go_test(go_available, f"TestFindNotebooks/{case}")


# ── picker behavior ────────────────────────────────────────────────────────────

@pytest.mark.parametrize("case", [
    "no_notebooks_shows_picker",
    "one_notebook_skips_picker",
    "two_notebooks_shows_picker",
    "only_non-notebook_files_shows_picker",
    "three_notebooks_shows_picker",
])
def test_picker_behavior(go_available, case):
    _go_test(go_available, f"TestPickerBehavior/{case}")


# ── selectRunner ───────────────────────────────────────────────────────────────

@pytest.mark.parametrize("case", [
    "ipynb_uses_juv",
    "py_with_marimo_dep_uses_marimo",
    "py_without_marimo_uses_uv_run",
    "py_with_empty_content_uses_uv_run",
    "py_with_marimo_in_comment_string_uses_marimo",
])
def test_select_runner(go_available, case):
    _go_test(go_available, f"TestSelectRunner/{case}")

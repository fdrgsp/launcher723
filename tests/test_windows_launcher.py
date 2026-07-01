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


# ── selectRunner ───────────────────────────────────────────────────────────────


@pytest.mark.parametrize(
    "case",
    [
        "ipynb_uses_juv",
        "ipynb_with_juv-mode_exec",
        "ipynb_with_juv-mode_run_explicit",
        "ipynb_with_no_pyrunner_section_defaults_to_run",
        "py_with_marimo_dep_edit_mode",
        "py_with_marimo_dep_run_mode",
        "py_without_marimo_uses_uv_run",
        "py_with_empty_content_uses_uv_run",
        "py_with_marimo_version_spec_edit_mode",
        "py_with_single-quoted_marimo_edit_mode",
        "py_with_unrelated_marimo_mention_uses_uv_run",
        "py_with_marimo_dep_no_pyrunner_section_defaults_to_edit",
    ],
)
def test_select_runner(go_available, case):
    _go_test(go_available, f"TestSelectRunner/{case}")


# ── TestMarimoMode ───────────────────────────────────────────────────────────────


@pytest.mark.parametrize(
    "case",
    [
        "no_script_block",
        "run_mode",
        "edit_mode",
        "single-quoted_run_mode",
        "no_pyrunner_section",
        "section_without_marimo-mode",
        "marimo-mode_after_other_keys",
    ],
)
def test_marimo_mode(go_available, case):
    _go_test(go_available, f"TestMarimoMode/{case}")


# ── TestJuvMode ──────────────────────────────────────────────────────────────────


@pytest.mark.parametrize(
    "case",
    [
        "no_script_block",
        "run_mode",
        "exec_mode",
        "single-quoted_exec_mode",
        "no_pyrunner_section",
        "section_without_juv-mode",
        "juv-mode_after_other_keys",
        "trailing_inline_comment_with_another_quoted_mode_value",
        "malformed_json_defaults_to_empty",
        "invalid_json_returns_empty",
    ],
)
def test_juv_mode(go_available, case):
    _go_test(go_available, f"TestJuvMode/{case}")

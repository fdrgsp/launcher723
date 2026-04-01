"""Tests for macos/launch — calls the real bash functions via subprocess."""

import os
import subprocess

import pytest

LAUNCH_SCRIPT = os.path.join(os.path.dirname(__file__), "..", "macos", "launch")


def _bash(cmd: str) -> str:
    result = subprocess.run(["bash", "-c", cmd], capture_output=True, text=True)
    return result.stdout.strip()


def _call(func: str, *args: str) -> str:
    quoted = " ".join(f'"{a}"' for a in args)
    return _bash(f'source "{LAUNCH_SCRIPT}" && {func} {quoted}')


# ── find_sole_notebook ─────────────────────────────────────────────────────────

find_sole_notebook_cases = [
    ("empty dir", [], None),
    ("single ipynb", ["notebook.ipynb"], "notebook.ipynb"),
    ("single py", ["script.py"], "script.py"),
    ("two notebooks", ["a.ipynb", "b.py"], None),
    ("three notebooks", ["a.ipynb", "b.ipynb", "c.py"], None),
    ("non-notebook files only", ["readme.txt", "data.csv"], None),
    ("notebook among non-notebooks", ["nb.ipynb", "readme.txt"], "nb.ipynb"),
]


@pytest.mark.parametrize("desc,files,expected_basename", find_sole_notebook_cases)
def test_find_sole_notebook(tmp_path, desc, files, expected_basename):
    for f in files:
        (tmp_path / f).touch()
    result = _call("find_sole_notebook", str(tmp_path))
    actual = os.path.basename(result) if result else None
    assert actual == expected_basename


# ── select_runner ──────────────────────────────────────────────────────────────

select_runner_cases = [
    ("ipynb uses juv", "notebook.ipynb", "", "uvx juv run"),
    ("py with marimo dep", "nb.py", '# dependencies = [\n#   "marimo",\n# ]', "uvx marimo edit --sandbox"),
    ("py without marimo", "script.py", '# dependencies = [\n#   "numpy",\n# ]', "uv run"),
    ("py empty content", "script.py", "", "uv run"),
]


@pytest.mark.parametrize("desc,filename,content,expected", select_runner_cases)
def test_select_runner(tmp_path, desc, filename, content, expected):
    path = tmp_path / filename
    path.write_text(content)
    actual = _call("select_runner", str(path))
    assert actual == expected


# ── is_translocated ────────────────────────────────────────────────────────────

is_translocated_cases = [
    ("/private/var/folders/xy/abc/AppTranslocation/UUID/d/launcher723.app", True),
    ("/Users/test/Desktop/untitled folder", False),
    ("/Users/test", False),
    ("/private/var/folders/xy/abc", False),
]


@pytest.mark.parametrize("path,expected", is_translocated_cases)
def test_is_translocated(path, expected):
    cmd = f'source "{LAUNCH_SCRIPT}" && is_translocated "{path}" && echo true || echo false'
    actual = _bash(cmd) == "true"
    assert actual == expected


# ── picker behavior ────────────────────────────────────────────────────────────

picker_cases = [
    ("no notebooks", [], True),
    ("one ipynb", ["nb.ipynb"], False),
    ("one py", ["script.py"], False),
    ("two notebooks", ["a.ipynb", "b.py"], True),
    ("only txt", ["readme.txt"], True),
    ("three notebooks", ["a.ipynb", "b.ipynb", "c.py"], True),
]


@pytest.mark.parametrize("desc,files,expect_picker", picker_cases)
def test_picker_behavior(tmp_path, desc, files, expect_picker):
    for f in files:
        (tmp_path / f).touch()
    result = _call("find_sole_notebook", str(tmp_path))
    assert (not result) == expect_picker

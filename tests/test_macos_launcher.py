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


# ── select_runner ──────────────────────────────────────────────────────────────

select_runner_cases = [
    ("ipynb uses juv", "notebook.ipynb", "", "uvx juv run"),
    (
        "py with marimo dep",
        "nb.py",
        '# dependencies = [\n#   "marimo",\n# ]',
        "uvx marimo edit --sandbox",
    ),
    (
        "py without marimo",
        "script.py",
        '# dependencies = [\n#   "numpy",\n# ]',
        "uv run",
    ),
    ("py empty content", "script.py", "", "uv run"),
    (
        "py with marimo version spec",
        "nb.py",
        '# dependencies = [\n#   "marimo>=0.1",\n# ]',
        "uvx marimo edit --sandbox",
    ),
    (
        "py with single-quoted marimo",
        "nb.py",
        "# dependencies = [\n#   'marimo',\n# ]",
        "uvx marimo edit --sandbox",
    ),
    (
        "py with unrelated marimo mention",
        "script.py",
        "# this is not marimo_extra related",
        "uv run",
    ),
]


@pytest.mark.parametrize("desc,filename,content,expected", select_runner_cases)
def test_select_runner(tmp_path, desc, filename, content, expected):
    path = tmp_path / filename
    path.write_text(content)
    actual = _call("select_runner", str(path))
    assert actual == expected

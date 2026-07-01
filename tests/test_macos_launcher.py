"""Tests for macos/launch.sh — calls the real bash functions via subprocess."""

import json
import os
import subprocess

import pytest

LAUNCH_SCRIPT = os.path.join(os.path.dirname(__file__), "..", "macos", "launch.sh")


def _bash(cmd: str) -> str:
    result = subprocess.run(["bash", "-c", cmd], capture_output=True, text=True)
    return result.stdout.strip()


def _call(func: str, *args: str) -> str:
    quoted = " ".join(f'"{a}"' for a in args)
    return _bash(f'source "{LAUNCH_SCRIPT}" && {func} {quoted}')


def _ipynb(header_lines: list[str]) -> str:
    """Build a minimal .ipynb JSON string with header_lines as the source of
    the hidden PEP 723 metadata cell — mirrors what `juv add` produces."""
    nb = {
        "cells": [
            {
                "cell_type": "code",
                "execution_count": None,
                "id": "meta",
                "metadata": {"jupyter": {"source_hidden": True}},
                "outputs": [],
                "source": header_lines,
            }
        ],
        "metadata": {},
        "nbformat": 4,
        "nbformat_minor": 5,
    }
    return json.dumps(nb, indent=1)


# ── select_runner ──────────────────────────────────────────────────────────────

select_runner_cases = [
    ("ipynb uses juv", "notebook.ipynb", "", "uvx juv run"),
    (
        "ipynb with juv-mode exec",
        "notebook.ipynb",
        _ipynb(["# /// script\n", "# [pyrunner]\n", '# juv-mode = "exec"\n', "# ///"]),
        "uvx juv exec",
    ),
    (
        "ipynb with juv-mode run explicit",
        "notebook.ipynb",
        _ipynb(["# /// script\n", "# [pyrunner]\n", '# juv-mode = "run"\n', "# ///"]),
        "uvx juv run",
    ),
    (
        "ipynb with no pyrunner section defaults to run",
        "notebook.ipynb",
        _ipynb(
            [
                "# /// script\n",
                "# dependencies = [\n",
                '#   "numpy",\n',
                "# ]\n",
                "# ///",
            ]
        ),
        "uvx juv run",
    ),
    (
        "py with marimo dep edit mode",
        "nb.py",
        '# /// script\n# dependencies = [\n#   "marimo",\n# ]\n#\n'
        '# [pyrunner]\n# marimo-mode = "edit"\n# ///\n',
        "uvx marimo edit --sandbox",
    ),
    (
        "py with marimo dep run mode",
        "nb.py",
        '# /// script\n# dependencies = [\n#   "marimo",\n# ]\n#\n'
        '# [pyrunner]\n# marimo-mode = "run"\n# ///\n',
        "uvx marimo run --sandbox",
    ),
    (
        "py without marimo",
        "script.py",
        '# dependencies = [\n#   "numpy",\n# ]',
        "uv run",
    ),
    ("py empty content", "script.py", "", "uv run"),
    (
        "py with marimo version spec edit mode",
        "nb.py",
        '# /// script\n# dependencies = [\n#   "marimo>=0.1",\n# ]\n#\n'
        '# [pyrunner]\n# marimo-mode = "edit"\n# ///\n',
        "uvx marimo edit --sandbox",
    ),
    (
        "py with single-quoted marimo edit mode",
        "nb.py",
        "# /// script\n# dependencies = [\n#   'marimo',\n# ]\n#\n"
        '# [pyrunner]\n# marimo-mode = "edit"\n# ///\n',
        "uvx marimo edit --sandbox",
    ),
    (
        "py with unrelated marimo mention",
        "script.py",
        "# this is not marimo_extra related",
        "uv run",
    ),
    (
        "py with marimo dep no pyrunner section defaults to edit",
        "nb.py",
        '# /// script\n# dependencies = [\n#   "marimo",\n# ]\n# ///\n',
        "uvx marimo edit --sandbox",
    ),
]


@pytest.mark.parametrize("desc,filename,content,expected", select_runner_cases)
def test_select_runner(tmp_path, desc, filename, content, expected):
    path = tmp_path / filename
    path.write_text(content)
    actual = _call("select_runner", str(path))
    assert actual == expected


# ── marimo_mode ────────────────────────────────────────────────────────────────

marimo_mode_cases = [
    ("no script block", '# dependencies = [\n#   "marimo",\n# ]', ""),
    ("run mode", '# /// script\n# [pyrunner]\n# marimo-mode = "run"\n# ///\n', "run"),
    (
        "edit mode",
        '# /// script\n# [pyrunner]\n# marimo-mode = "edit"\n# ///\n',
        "edit",
    ),
    (
        "single-quoted run mode",
        "# /// script\n# [pyrunner]\n# marimo-mode = 'run'\n# ///\n",
        "run",
    ),
    (
        "no pyrunner section",
        '# /// script\n# dependencies = [\n#   "marimo",\n# ]\n# ///\n',
        "",
    ),
    (
        "section without marimo-mode",
        '# /// script\n# [pyrunner]\n# other_key = "value"\n# ///\n',
        "",
    ),
    (
        "marimo-mode after other keys",
        '# /// script\n# [pyrunner]\n# other = "x"\n# marimo-mode = "run"\n# ///\n',
        "run",
    ),
]


@pytest.mark.parametrize("desc,content,expected", marimo_mode_cases)
def test_marimo_mode(tmp_path, desc, content, expected):
    path = tmp_path / "nb.py"
    path.write_text(content)
    actual = _call("marimo_mode", str(path))
    assert actual == expected


# ── juv_mode ───────────────────────────────────────────────────────────────────

juv_mode_cases = [
    ("no script block", ["print('hi')\n"], ""),
    (
        "run mode",
        ["# /// script\n", "# [pyrunner]\n", '# juv-mode = "run"\n', "# ///"],
        "run",
    ),
    (
        "exec mode",
        ["# /// script\n", "# [pyrunner]\n", '# juv-mode = "exec"\n', "# ///"],
        "exec",
    ),
    (
        "single-quoted exec mode",
        ["# /// script\n", "# [pyrunner]\n", "# juv-mode = 'exec'\n", "# ///"],
        "exec",
    ),
    (
        "no pyrunner section",
        ["# /// script\n", "# dependencies = [\n", '#   "numpy",\n', "# ]\n", "# ///"],
        "",
    ),
    (
        "section without juv-mode",
        ["# /// script\n", "# [pyrunner]\n", '# other_key = "value"\n', "# ///"],
        "",
    ),
    (
        "juv-mode after other keys",
        [
            "# /// script\n",
            "# [pyrunner]\n",
            '# other = "x"\n',
            '# juv-mode = "exec"\n',
            "# ///",
        ],
        "exec",
    ),
    (
        "escaped quotes elsewhere in the block don't confuse the scan",
        [
            "# /// script\n",
            "# dependencies = [\n",
            '#   "numpy>=1.26",\n',
            "# ]\n",
            "#\n",
            "# [pyrunner]\n",
            '# juv-mode = "exec"\n',
            "# ///",
        ],
        "exec",
    ),
    (
        "trailing inline comment with another quoted mode value",
        [
            "# /// script\n",
            "# [pyrunner]\n",
            '# juv-mode = "exec"  # or "run" (default)\n',
            "# ///",
        ],
        "exec",
    ),
]


@pytest.mark.parametrize("desc,header_lines,expected", juv_mode_cases)
def test_juv_mode(tmp_path, desc, header_lines, expected):
    path = tmp_path / "nb.ipynb"
    path.write_text(_ipynb(header_lines))
    actual = _call("juv_mode", str(path))
    assert actual == expected

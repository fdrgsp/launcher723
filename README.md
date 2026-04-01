<h1><img src="icon.png" width="100" alt="launcher723 icon"/><br/>launcher723</h1>

A standalone desktop launcher for **Jupyter** and **marimo** notebooks, and plain Python scripts — double-click to run any `.ipynb` or `.py` file that has inline dependencies, no Python installation required.

Files must declare their dependencies using the [PEP 723](https://peps.python.org/pep-0723/) inline script metadata format. The launcher uses [juv](https://github.com/manzt/juv) + [uv](https://docs.astral.sh/uv/) for `.ipynb` files, [marimo](https://marimo.io/) + [uv](https://docs.astral.sh/uv/) for `.py` files that declare `marimo` as a dependency, and plain `uv run` for other `.py` scripts.

## Download

| macOS                                                                                                              | Windows                                                                                                                  |
|--------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| [Download .app](https://github.com/fdrgsp/launcher723/releases/download/latest/launcher723-latest-mac.zip)         | [Download .exe](https://github.com/fdrgsp/launcher723/releases/download/latest/launcher723-latest-windows.zip)           |

## How it works

1. Download and unzip the file for your platform
2. Double-click the executable (`.app` on macOS, `.exe` on Windows)
3. A file picker appears — select a `.ipynb` or `.py` file
4. The file opens in your browser (notebooks) or terminal (scripts)

**First run only:** `uv` and all packages are downloaded and cached automatically. Subsequent runs are fast.

## Preparing your file

### Jupyter notebooks (`.ipynb`)

The launcher runs `.ipynb` files using `uvx juv run`, which requires each notebook to declare its dependencies inline using [PEP 723](https://peps.python.org/pep-0723/) script metadata.

#### Option 1: Use `juv add` (recommended)

Install [uv](https://docs.astral.sh/uv/getting-started/installation/), then add dependencies to your notebook:

```bash
uvx juv add my_notebook.ipynb numpy pandas matplotlib
```

This writes a hidden `# /// script` cell into the notebook. You can verify it works locally:

```bash
uvx juv run my_notebook.ipynb
```

#### Option 2: Add the metadata manually

Add a code cell at the top of your notebook with:

```python
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "numpy>=1.26",
#   "pandas>=2.0",
#   "matplotlib>=3.8",
# ]
# ///
```

This cell tells `juv` which packages to install in the isolated environment.

### Python scripts (`.py`)

The launcher inspects each `.py` file for [PEP 723](https://peps.python.org/pep-0723/) inline script metadata and chooses the right runner automatically:

- If `marimo` is listed as a dependency → runs with `uvx marimo edit --sandbox`
- Otherwise → runs with `uv run`

#### marimo notebooks

```python
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "marimo",
#   "numpy>=1.26",
#   "pandas>=2.0",
# ]
# ///
```

You can verify it works locally:

```bash
uvx marimo edit my_notebook.py
```

#### Plain Python scripts

```python
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "numpy>=1.26",
#   "pandas>=2.0",
# ]
# ///
```

You can verify it works locally:

```bash
uv run my_script.py
```

## Platform notes

**macOS Gatekeeper warning:** on first launch macOS may show a popup with only "Delete" or "Done" as options and block the app from opening. If this happens:

1. Open **System Settings → Privacy & Security**
2. Scroll to the bottom of the page
3. You will see a message about the app being blocked — click **"Open Anyway"**

**Windows SmartScreen warning:** on first launch Windows may show a warning that the app is from an unknown publisher. If this happens:

1. Click **"More info"** on the warning dialog
2. Click **"Run anyway"**

## What users need

- An internet connection on first run (to fetch `uv` and packages — cached locally after that)
- Nothing else — no Python, no conda, no pip

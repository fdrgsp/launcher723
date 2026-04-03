# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "numpy>=1.26",
#   "matplotlib>=3.8",
#   "marimo>=0.22.0",
# ]
#
# [pyrunner]
# marimo-mode = "run"
# ///

import marimo

__generated_with = "0.22.0"
app = marimo.App(width="medium")


@app.cell
def _():
    import marimo as mo

    return (mo,)


@app.cell(hide_code=True)
def _(mo):
    mo.md(r"""
    # Example Notebook

    Replace this notebook with your own. Dependencies are declared in the hidden cell
    above. Use `juv add notebook.ipynb <package>` to add more or manually edit the cell.
    """)
    return


@app.cell
def _(mo):
    freq = mo.ui.slider(1, 10, value=1, step=0.5, label="Frequency")
    freq
    return (freq,)


@app.cell
def _(freq):
    import matplotlib.pyplot as plt
    import numpy as np

    x = np.linspace(0, 2 * np.pi, 300)
    _fig, ax = plt.subplots()
    ax.plot(x, np.sin(freq.value * x))
    ax.set_title("Hello from pyrunner!")
    return


@app.cell
def _():
    return


if __name__ == "__main__":
    app.run()

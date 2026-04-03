# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "numpy>=1.26",
#   "matplotlib>=3.8",
#   "marimo>=0.22.0"
# ]
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

    Replace this notebook with your own. Dependencies are declared in the hidden cell above — use `juv add notebook.ipynb <package>` to add more.
    """)
    return


@app.cell
def _():
    import numpy as np
    import matplotlib.pyplot as plt

    x = np.linspace(0, 2 * np.pi, 300)
    plt.plot(x, np.sin(x))
    plt.title("Hello from juv + PEP 723")

    plt.show()
    return


@app.cell
def _():
    print("Done!")
    return



if __name__ == "__main__":
    app.run()

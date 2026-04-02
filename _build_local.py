# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""Build pyrunner packages locally for testing.

Usage:
    uv run _build_local.py                        # build for current platform
    uv run _build_local.py -o                     # build and open the .app (macOS)
    uv run _build_local.py -t all                 # build both mac and windows
"""

import argparse
import platform
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
DIST = ROOT / "dist"


def build_mac() -> Path:
    app = DIST / "pyrunner.app" / "Contents"
    for d in (app / "MacOS", app / "Resources"):
        d.mkdir(parents=True, exist_ok=True)

    shutil.copy2(ROOT / "macos" / "Info.plist", app / "Info.plist")
    shutil.copy2(ROOT / "macos" / "icon.icns", app / "Resources" / "icon.icns")

    # Compile universal binary stub so LaunchServices sees a native executable
    # (avoids Rosetta prompt on Apple Silicon)
    launch_bin = app / "MacOS" / "launch"
    subprocess.run(
        [
            "cc",
            "-arch",
            "arm64",
            "-arch",
            "x86_64",
            "-o",
            str(launch_bin),
            str(ROOT / "macos" / "launch.c"),
        ],
        check=True,
    )
    launch_bin.chmod(0o755)

    shutil.copy2(ROOT / "macos" / "launch.sh", app / "MacOS" / "launch.sh")
    (app / "MacOS" / "launch.sh").chmod(0o755)

    # Remove quarantine attribute to avoid Gatekeeper warning
    subprocess.run(["xattr", "-cr", str(app.parent)], check=False)

    print(f"Built: {app.parent}")
    return app.parent


def build_windows(arch: str = "amd64") -> Path:
    exe_dir = DIST / f"pyrunner-windows-{arch}"
    exe_dir.mkdir(parents=True, exist_ok=True)

    if shutil.which("go") is None:
        print(
            "Error: Go is not installed (needed to build the Windows .exe).",
            file=sys.stderr,
        )
        sys.exit(1)

    subprocess.run(
        ["go", "build", "-ldflags", "-s -w", "-o", str(exe_dir / "pyrunner.exe"), "."],
        cwd=ROOT / "windows",
        check=True,
        env={**__import__("os").environ, "GOOS": "windows", "GOARCH": arch},
    )

    print(f"Built: {exe_dir / 'pyrunner.exe'}")
    return exe_dir


def main() -> None:
    parser = argparse.ArgumentParser(description="Build pyrunner locally for testing.")
    parser.add_argument(
        "-o",
        "--open",
        action="store_true",
        help="Open the .app after building (macOS only)",
    )
    parser.add_argument(
        "-t",
        "--target",
        choices=["mac", "windows", "all"],
        default=None,
        help="Target platform (default: current platform)",
    )
    args = parser.parse_args()

    system = platform.system()
    target = args.target or ("mac" if system == "Darwin" else "windows")
    targets = ["mac", "windows"] if target == "all" else [target]

    for t in targets:
        if t == "mac":
            result = build_mac()
            if args.open and system == "Darwin":
                subprocess.run(["open", str(result)])
        elif t == "windows":
            for arch in ("amd64", "arm64"):
                build_windows(arch)
            if args.open:
                print("Note: --open is only supported on macOS.")


if __name__ == "__main__":
    main()

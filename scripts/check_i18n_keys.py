#!/usr/bin/env python3
"""check_i18n_keys.py

Utility to verify that every i18n lookup present in the Go source code
exists in the default English locale file (locales/en.yml).

It recursively walks the Go code base (default: ./alita), extracts string
literals passed to `GetString` and `GetStringSlice` calls, and checks that
corresponding keys are defined in the YAML locale.

Usage (from repository root):
    python scripts/check_i18n_keys.py

Optional flags:
    -c / --code-dir   Path to the directory with Go source files (default: alita)
    -l / --locale     Path to the YAML locale file (default: locales/en.yml)

Requires: PyYAML (install with `pip install pyyaml`).
"""
from __future__ import annotations

import argparse
import os
import re
import sys
from collections import defaultdict
from typing import Dict, List, Set

try:
    import yaml  # type: ignore
except ImportError as exc:
    sys.exit("PyYAML is required for this script: pip install pyyaml")

# ---------------------------------------------------------------------------
# Helpers for YAML key extraction
# ---------------------------------------------------------------------------


def _collect_yaml_keys(data: object, prefix: str = "") -> Set[str]:
    """Recursively collect dotted key paths from a nested dict loaded from YAML."""
    keys: Set[str] = set()

    if isinstance(data, dict):
        for k, v in data.items():
            new_prefix = f"{prefix}.{k}" if prefix else str(k)
            keys.update(_collect_yaml_keys(v, new_prefix))
    else:
        # Leaf node – register the accumulated prefix as a full key path
        if prefix:
            keys.add(prefix)
    return keys


def load_locale_keys(locale_path: str) -> Set[str]:
    """Load all dotted key paths from the given YAML locale file."""
    with open(locale_path, "r", encoding="utf-8") as fh:
        yaml_data = yaml.safe_load(fh) or {}
    return _collect_yaml_keys(yaml_data)


# ---------------------------------------------------------------------------
# Helpers for parsing Go source files
# ---------------------------------------------------------------------------

# Regex to capture string literals passed to .GetString("foo.bar") or
# .GetStringSlice("foo.bar") – keeps the key in group 1.
GETSTRING_RE = re.compile(r"\.GetString(?:Slice)?\(\s*\"([^\"]+)\"\s*\)")


def scan_go_files(code_dir: str) -> Dict[str, List[str]]:
    """Return a mapping of used i18n keys to a list of locations where they're used."""
    used: Dict[str, List[str]] = defaultdict(list)

    for root, _dirs, files in os.walk(code_dir):
        for filename in files:
            if not filename.endswith(".go"):
                continue
            path = os.path.join(root, filename)
            try:
                with open(path, "r", encoding="utf-8", errors="ignore") as fh:
                    for lineno, line in enumerate(fh, 1):
                        for match in GETSTRING_RE.finditer(line):
                            key = match.group(1)
                            used[key].append(f"{path}:{lineno}")
            except (OSError, UnicodeError) as exc:
                print(f"[WARN] Could not read {path}: {exc}", file=sys.stderr)
    return used


# ---------------------------------------------------------------------------
# Main logic
# ---------------------------------------------------------------------------


def main() -> None:
    parser = argparse.ArgumentParser(description="Detect missing i18n keys.")
    parser.add_argument(
        "-c",
        "--code-dir",
        default="alita",
        help="Directory containing Go source files to scan (default: alita)",
    )
    parser.add_argument(
        "-l",
        "--locale",
        default="locales/en.yml",
        help="Path to the English locale YAML file (default: locales/en.yml)",
    )

    args = parser.parse_args()

    if not os.path.isfile(args.locale):
        sys.exit(f"Locale file not found: {args.locale}")

    if not os.path.isdir(args.code_dir):
        sys.exit(f"Code directory not found: {args.code_dir}")

    en_keys = load_locale_keys(args.locale)
    used_keys_map = scan_go_files(args.code_dir)

    missing_keys = []
    for k in used_keys_map:
        # direct match
        if k in en_keys:
            continue
        # fallback try with "strings." prefix
        if not k.startswith("strings.") and f"strings.{k}" in en_keys:
            continue
        missing_keys.append(k)

    if not missing_keys:
        print("✅ All i18n keys used in the code are present in the English locale.")
        return

    print(f"❌ Missing i18n keys detected ({len(missing_keys)}):\n")
    for key in sorted(missing_keys):
        print(f" - {key}")
        for location in used_keys_map[key]:
            print(f"     used at: {location}")

    # Exit with non-zero status so this can be used in CI.
    sys.exit(1)


if __name__ == "__main__":
    main()

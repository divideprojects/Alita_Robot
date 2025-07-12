#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Utility to verify that every i18n lookup present in the Go source code
exists in the default English locale file (locales/en.yml).

It recursively walks the Go code base (default: ./alita), extracts string
literals passed to `GetString` and `GetStringSlice` calls, and checks that
corresponding keys are defined in the YAML locale.

Usage (from repository root):
    python scripts/check_code_keys.py

Optional flags:
    -c / --code-dir   Path to the directory with Go source files (default: alita)
    -l / --locale     Path to the YAML locale file (default: locales/en.yml)

Requires: ruamel.yaml (install with `pip install ruamel.yaml`).
"""
from __future__ import annotations

import argparse
import os
import re
import sys
from collections import defaultdict
from typing import Dict, List, Set

try:
    from ruamel.yaml import YAML
except ImportError:
    sys.exit("ruamel.yaml is required for this script: pip install ruamel.yaml")

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
    yaml = YAML(typ="safe")  # typ='safe' is equivalent to safe_load
    with open(locale_path, "r", encoding="utf-8") as fh:
        yaml_data = yaml.load(fh) or {}
    return _collect_yaml_keys(yaml_data)


# ---------------------------------------------------------------------------
# Helpers for parsing Go source files
# ---------------------------------------------------------------------------

# Regex to capture string literals passed to .GetString("foo.bar") or
# .GetStringSlice("foo.bar") – keeps the key in group 1.
GETSTRING_RE = re.compile(r"\.GetString(?:Slice)?\(\s*\"([^\"]+)\"\s*\)")

# Regex to capture GetStringWithError usage for production readiness analysis
GETSTRING_WITH_ERROR_RE = re.compile(r"\.GetStringWithError\(\s*\"([^\"]+)\"\s*\)")


def scan_go_files(code_dir: str):
    """Return mappings of used i18n keys to locations, and production readiness analysis."""
    used: Dict[str, List[str]] = defaultdict(list)
    with_error: Dict[str, List[str]] = defaultdict(list)
    critical_paths: Dict[str, List[str]] = defaultdict(list)

    # Critical modules that should use GetStringWithError for user-facing messages
    critical_modules = {
        "warns.go",
        "admin.go",
        "bans.go",
        "mute.go",
        "help.go",
        "filters.go",
        "notes.go",
    }

    for root, _dirs, files in os.walk(code_dir):
        for filename in files:
            if not filename.endswith(".go"):
                continue
            path = os.path.join(root, filename)
            is_critical = any(
                critical_mod in filename for critical_mod in critical_modules
            )

            try:
                with open(path, "r", encoding="utf-8", errors="ignore") as fh:
                    for lineno, line in enumerate(fh, 1):
                        # Check for regular GetString usage
                        for match in GETSTRING_RE.finditer(line):
                            key = match.group(1)
                            location = f"{path}:{lineno}"
                            used[key].append(location)

                            # Mark critical paths that should use GetStringWithError
                            if is_critical and any(
                                pattern in key
                                for pattern in [
                                    ".success",
                                    ".error",
                                    ".warn",
                                    ".ban",
                                    ".mute",
                                    ".kick",
                                ]
                            ):
                                critical_paths[key].append(location)

                        # Check for GetStringWithError usage
                        for match in GETSTRING_WITH_ERROR_RE.finditer(line):
                            key = match.group(1)
                            with_error[key].append(f"{path}:{lineno}")

            except (OSError, UnicodeError) as exc:
                print(f"[WARN] Could not read {path}: {exc}", file=sys.stderr)

    return used, with_error, critical_paths


# ---------------------------------------------------------------------------
# Main logic
# ---------------------------------------------------------------------------


def main() -> None:
    parser = argparse.ArgumentParser(description="Detect missing i18n keys in Go code.")
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
    used_keys_map, with_error_keys, critical_paths = scan_go_files(args.code_dir)

    missing_keys = []
    for k in used_keys_map:
        # direct match
        if k in en_keys:
            continue
        # fallback try with "strings." prefix
        if not k.startswith("strings.") and f"strings.{k}" in en_keys:
            continue
        missing_keys.append(k)

    # Production readiness analysis
    critical_without_error = []
    for key in critical_paths:
        if key not in with_error_keys:
            critical_without_error.append(key)

    # Report results
    has_issues = False

    if missing_keys:
        print(f"❌ Missing i18n keys detected ({len(missing_keys)}):\n")
        for key in sorted(missing_keys):
            print(f" - {key}")
            for location in used_keys_map[key]:
                print(f"     used at: {location}")
        print()
        has_issues = True

    if critical_without_error:
        print(
            f"⚠️  Critical paths not using GetStringWithError ({len(critical_without_error)}):"
        )
        print(
            "   These user-facing messages should use GetStringWithError for production safety:\n"
        )
        for key in sorted(critical_without_error):
            print(f" - {key}")
            for location in critical_paths[key]:
                print(f"     used at: {location}")
        print()
        has_issues = True

    # Summary statistics
    total_keys = len(used_keys_map)
    with_error_count = len(with_error_keys)
    critical_count = len(critical_paths)

    print("📊 Translation Statistics:")
    print(f"   Total i18n keys used: {total_keys}")
    print(
        f"   Keys using GetStringWithError: {with_error_count} ({with_error_count/total_keys*100:.1f}%)"
    )
    print(f"   Critical user-facing keys: {critical_count}")
    print(
        f"   Critical keys with error handling: {critical_count - len(critical_without_error)}/{critical_count}"
    )

    if not has_issues:
        print("\n✅ All i18n keys are present and production-ready!")
        return

    # Exit with non-zero status so this can be used in CI.
    sys.exit(1)


if __name__ == "__main__":
    main()

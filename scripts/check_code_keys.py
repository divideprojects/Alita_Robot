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
# .GetStringSlice("foo.bar") – requires "strings." prefix for consistency.
GETSTRING_RE = re.compile(r"\.GetString(?:Slice)?\(\s*\"(strings\.[^\"]+)\"\s*\)")

# Regex to capture GetStringWithError usage for production readiness analysis
GETSTRING_WITH_ERROR_RE = re.compile(
    r"\.GetStringWithError\(\s*\"(strings\.[^\"]+)\"\s*\)"
)

# Regex to capture any GetString calls that don't start with "strings." prefix (invalid)
INVALID_GETSTRING_RE = re.compile(
    r"\.GetString(?:Slice|WithError)?\(\s*\"([^s][^\"]*|s[^t][^\"]*|st[^r][^\"]*|str[^i][^\"]*|stri[^n][^\"]*|strin[^g][^\"]*|string[^s][^\"]*|strings[^\.][^\"]*)\"\s*\)"
)

# Special cases that are allowed (main namespace keys)
ALLOWED_NON_STRINGS_KEYS = {
    "main.language_name",
    "main.language_flag",
    "main.lang_sample",
}

# Test keys that should be ignored for missing key validation
TEST_KEYS = {
    "strings.test.key",
    "strings.test.list",
    "nonexistent.key",
    "nonexistent.list",
    "test.key",
    "strings.welcome.message",  # Documentation example
    "strings.errors.not_found",  # Documentation example
    "welcome.message",  # Documentation example
    "errors.not_found",  # Documentation example
}


def scan_go_files(code_dir: str):
    """Return mappings of used i18n keys to locations, and production readiness analysis."""
    used: Dict[str, List[str]] = defaultdict(list)
    with_error: Dict[str, List[str]] = defaultdict(list)
    critical_paths: Dict[str, List[str]] = defaultdict(list)
    invalid_keys: Dict[str, List[str]] = defaultdict(list)

    # Critical modules that should use GetStringWithError for user-facing messages
    critical_modules = {
        "warns.go",
        "admin.go",
        "bans.go",
        "mute.go",
        "help.go",
        "filters.go",
        "notes.go",
        "greetings.go",
    }

    for root, dirs, files in os.walk(code_dir):
        for file in files:
            if not file.endswith(".go"):
                continue

            # Skip test files for key validation
            is_test_file = file.endswith("_test.go")

            filepath = os.path.join(root, file)
            try:
                with open(filepath, "r", encoding="utf-8") as f:
                    content = f.read()
            except UnicodeDecodeError:
                continue

            lines = content.split("\n")
            for line_num, line in enumerate(lines, 1):
                # Skip comment lines
                stripped_line = line.strip()
                if (
                    stripped_line.startswith("//")
                    or stripped_line.startswith("*")
                    or stripped_line.startswith("/*")
                ):
                    continue

                # Find valid GetString calls with "strings." prefix
                for match in GETSTRING_RE.finditer(line):
                    key = match.group(1)
                    location = f"{filepath}:{line_num}"

                    # Skip test keys for missing key validation
                    if not is_test_file and key not in TEST_KEYS:
                        used[key].append(location)

                # Find GetStringWithError calls
                for match in GETSTRING_WITH_ERROR_RE.finditer(line):
                    key = match.group(1)
                    location = f"{filepath}:{line_num}"
                    with_error[key].append(location)

                    # Skip test keys for missing key validation
                    if not is_test_file and key not in TEST_KEYS:
                        used[key].append(location)

                    # Track critical paths
                    if any(critical_mod in file for critical_mod in critical_modules):
                        critical_paths[key].append(location)

                # Find invalid GetString calls (without "strings." prefix)
                for match in INVALID_GETSTRING_RE.finditer(line):
                    key = match.group(1)

                    # Skip allowed special cases and test keys
                    if key in ALLOWED_NON_STRINGS_KEYS or key in TEST_KEYS:
                        continue

                    invalid_keys[key].append(f"{filepath}:{line_num}")

                # Find regular GetString calls in critical modules for tracking
                for match in re.finditer(
                    r"\.GetString\(\s*\"(strings\.[^\"]+)\"\s*\)", line
                ):
                    key = match.group(1)
                    location = f"{filepath}:{line_num}"
                    if any(critical_mod in file for critical_mod in critical_modules):
                        critical_paths[key].append(location)

    return used, with_error, critical_paths, invalid_keys


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
    used_keys_map, with_error_keys, critical_paths, invalid_keys = scan_go_files(
        args.code_dir
    )

    missing_keys = []
    for k in used_keys_map:
        # All keys should start with "strings." prefix now
        if k not in en_keys:
            missing_keys.append(k)

    # Production readiness analysis
    critical_without_error = []
    for key in critical_paths:
        if key not in with_error_keys:
            critical_without_error.append(key)

    # Report results
    has_issues = False

    # Check for invalid key formats (keys without "strings." prefix)
    if invalid_keys:
        print(f"❌ Invalid i18n key format detected ({len(invalid_keys)}):")
        print("   All i18n keys MUST start with 'strings.' prefix for consistency.\n")
        for key in sorted(invalid_keys):
            print(f" - {key}")
            for location in invalid_keys[key]:
                print(f"     used at: {location}")
        print()
        print("💡 Fix: Update these keys to use 'strings.' prefix:")
        for key in sorted(invalid_keys):
            print(f'   tr.GetString("{key}") → tr.GetString("strings.{key}")')
        print()
        has_issues = True

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
    invalid_count = len(invalid_keys)

    print("📊 Translation Statistics:")
    print(f"   Total i18n keys used: {total_keys}")
    print(f"   Keys with valid 'strings.' prefix: {total_keys}")
    print(f"   Keys with invalid format: {invalid_count}")
    print(
        f"   Keys using GetStringWithError: {with_error_count} ({with_error_count/max(total_keys,1)*100:.1f}%)"
    )
    print(f"   Critical user-facing keys: {critical_count}")
    print(
        f"   Critical keys with error handling: {critical_count - len(critical_without_error)}/{critical_count}"
    )

    if not has_issues:
        print("\n✅ All i18n keys follow proper format and are production-ready!")
        return

    # Exit with non-zero status so this can be used in CI.
    sys.exit(1)


if __name__ == "__main__":
    main()

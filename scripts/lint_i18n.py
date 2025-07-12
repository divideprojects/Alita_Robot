#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
I18n Linting Script for Alita Robot

This script enforces strict i18n key format requirements:
- All i18n keys MUST start with "strings." prefix
- Critical user-facing messages should use GetStringWithError
- Provides clear error messages and fix suggestions

Usage:
    python3 scripts/lint_i18n.py
    python3 scripts/lint_i18n.py --fix-suggestions

Exit codes:
    0: All i18n usage follows proper format
    1: Format violations found
"""

import argparse
import os
import re
import sys
from pathlib import Path
from typing import List, Tuple, Dict


# Regex patterns for i18n validation
VALID_GETSTRING_RE = re.compile(
    r"\.GetString(?:Slice|WithError)?\(\s*\"(strings\.[^\"]+)\"\s*\)"
)
INVALID_GETSTRING_RE = re.compile(
    r"\.GetString(?:Slice|WithError)?\(\s*\"([^s][^\"]*|s[^t][^\"]*|st[^r][^\"]*|str[^i][^\"]*|stri[^n][^\"]*|strin[^g][^\"]*|string[^s][^\"]*|strings[^\.][^\"]*)\"\s*\)"
)
GETSTRING_WITH_ERROR_RE = re.compile(
    r"\.GetStringWithError\(\s*\"(strings\.[^\"]+)\"\s*\)"
)

# Special cases that are allowed (main namespace keys)
ALLOWED_NON_STRINGS_KEYS = {
    "main.language_name",
    "main.language_flag",
    "main.lang_sample",
}


class I18nLintError:
    """Represents an i18n linting error."""

    def __init__(
        self,
        filepath: str,
        line_num: int,
        error_type: str,
        key: str,
        suggestion: str = "",
    ):
        self.filepath = filepath
        self.line_num = line_num
        self.error_type = error_type
        self.key = key
        self.suggestion = suggestion

    def __str__(self):
        result = f"{self.filepath}:{self.line_num}: {self.error_type}: '{self.key}'"
        if self.suggestion:
            result += f"\n    Suggestion: {self.suggestion}"
        return result


def lint_file(filepath: Path) -> List[I18nLintError]:
    """Lint a single Go file for i18n format violations."""
    errors = []

    try:
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
    except UnicodeDecodeError:
        return errors

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

        # Find invalid GetString calls
        for match in INVALID_GETSTRING_RE.finditer(line):
            key = match.group(1)

            # Skip allowed special cases
            if key in ALLOWED_NON_STRINGS_KEYS:
                continue

            error = I18nLintError(
                filepath=str(filepath),
                line_num=line_num,
                error_type="INVALID_KEY_FORMAT",
                key=key,
                suggestion=f'Use "strings.{key}" instead of "{key}"',
            )
            errors.append(error)

    return errors


def lint_directory(code_dir: Path) -> List[I18nLintError]:
    """Lint all Go files in a directory for i18n format violations."""
    all_errors = []

    for go_file in code_dir.rglob("*.go"):
        # Skip vendor and test files
        if "vendor" in str(go_file) or "_test.go" in str(go_file):
            continue

        file_errors = lint_file(go_file)
        all_errors.extend(file_errors)

    return all_errors


def group_errors_by_type(errors: List[I18nLintError]) -> Dict[str, List[I18nLintError]]:
    """Group errors by their type for better reporting."""
    grouped = {}
    for error in errors:
        if error.error_type not in grouped:
            grouped[error.error_type] = []
        grouped[error.error_type].append(error)
    return grouped


def print_fix_suggestions(errors: List[I18nLintError]):
    """Print automated fix suggestions for common errors."""
    print("\n🔧 Automated Fix Suggestions:")
    print("=" * 50)

    # Group by file for easier fixing
    files_with_errors = {}
    for error in errors:
        if error.filepath not in files_with_errors:
            files_with_errors[error.filepath] = []
        files_with_errors[error.filepath].append(error)

    for filepath, file_errors in files_with_errors.items():
        print(f"\n📁 {filepath}:")
        for error in file_errors:
            if error.error_type == "INVALID_KEY_FORMAT":
                print(
                    f"  Line {error.line_num}: Replace '{error.key}' with 'strings.{error.key}'"
                )
                print(
                    f'    sed -i \'s/"{error.key}"/"strings.{error.key}"/g\' {filepath}'
                )


def main():
    parser = argparse.ArgumentParser(
        description="Lint i18n key usage for format compliance",
        formatter_class=argparse.RawTextHelpFormatter,
    )
    parser.add_argument(
        "--code-dir",
        default="alita",
        help="Directory containing Go source files to lint (default: alita)",
    )
    parser.add_argument(
        "--fix-suggestions", action="store_true", help="Show automated fix suggestions"
    )

    args = parser.parse_args()

    code_dir = Path(args.code_dir)
    if not code_dir.exists():
        print(f"❌ Error: Code directory '{code_dir}' does not exist", file=sys.stderr)
        sys.exit(1)

    print(f"🔍 Linting i18n usage in {code_dir}...")

    errors = lint_directory(code_dir)

    if not errors:
        print("✅ All i18n keys follow proper format requirements!")
        print("📊 Summary:")
        print("   - No format violations found")
        print("   - All keys use required 'strings.' prefix")
        return

    # Group and report errors
    grouped_errors = group_errors_by_type(errors)

    print(f"❌ Found {len(errors)} i18n format violation(s):")
    print()

    for error_type, type_errors in grouped_errors.items():
        print(f"📋 {error_type} ({len(type_errors)} violations):")
        for error in type_errors:
            print(f"  {error}")
        print()

    # Show fix suggestions if requested
    if args.fix_suggestions:
        print_fix_suggestions(errors)
    else:
        print("💡 Run with --fix-suggestions to see automated fix commands")

    print(f"\n📊 Summary:")
    print(f"   - Total violations: {len(errors)}")
    print(f"   - Files affected: {len(set(e.filepath for e in errors))}")
    print(f"   - Fix required: All i18n keys must start with 'strings.' prefix")

    sys.exit(1)


if __name__ == "__main__":
    main()

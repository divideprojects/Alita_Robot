#!/usr/bin/env python3
"""Extract unique i18n keys used in Go modules.

This script walks through the `alita/modules` directory, finds usages of
`GetString` and `GetStringWithError` in Go files, and prints a deduplicated,
sorted list of the keys that are requested from the i18n system.

Usage:
    python scripts/get_i18n_keys.py
"""

from __future__ import annotations

import ast
import re
import sys
from pathlib import Path
from typing import Set

# Third-party
try:
    import yaml  # type: ignore
except ImportError as exc:  # pragma: no cover – blanket except for clarity
    sys.stderr.write(
        "[error] Missing dependency 'pyyaml'. Install with: pip install pyyaml\n"
    )
    raise

# Repo root inferred two levels up from this script
REPO_ROOT = Path(__file__).resolve().parent.parent
# Paths
# Directory holding project Go source we want to inspect
SOURCE_DIR = REPO_ROOT / "alita"
LOCALE_FILE = REPO_ROOT / "locales" / "en.yml"

# Pattern finds `.GetString(` or `.GetStringWithError(` then captures until the matching ')' – but naive using first ')'. We'll rely on minimal dot usage and capture lazily up to `)`
CALL_PATTERN = re.compile(r"\.GetString(?:WithError)?\((.*?)\)", re.DOTALL)


def _evaluate_concat_expression(expr: str, module_title: str) -> str | None:
    """Attempt to evaluate a Go‐style string concatenation.

    Example expression: "strings." + m.moduleName + ".adminlist"

    We replace `m.moduleName` with the provided *module_title* wrapped in double
    quotes and then try to join literal parts split by '+' operators. If any
    part is not a literal string, we abort and return *None*.
    """

    # Replace placeholder with quoted module title
    expr = expr.replace("m.moduleName", f'"{module_title}"')

    parts = [p.strip() for p in expr.split("+")]
    lit_parts: list[str] = []

    for part in parts:
        # remove wrapping parentheses if any (simple case)
        part = part.lstrip("(").rstrip(")")

        if part.startswith("`") and part.endswith("`"):
            lit_parts.append(part[1:-1])
        elif part.startswith('"') and part.endswith('"'):
            try:
                lit_parts.append(ast.literal_eval(part))
            except Exception:
                lit_parts.append(part.strip('"'))
        else:
            # non-literal segment -> give up
            return None

    return "".join(lit_parts)


def extract_key(arg_source: str, module_title: str) -> str | None:
    """Attempt to evaluate the argument and return a clean key string.

    Returns:
        The extracted key if resolvable to a literal, otherwise the raw source
        (stripped) to provide at least some reference.
    """
    arg_source = arg_source.strip()

    # Keep only the first argument (before any comma)
    if "," in arg_source:
        arg_source = arg_source.split(",", 1)[0].strip()

    if not arg_source:
        return None

    # First, try evaluating concatenation expressions
    if "+" in arg_source or "m.moduleName" in arg_source:
        concat_eval = _evaluate_concat_expression(arg_source, module_title)
        if concat_eval:
            return concat_eval

    # Handle raw string literal using backticks
    if arg_source.startswith("`") and arg_source.endswith("`"):
        return arg_source.strip("`")

    # Handle quoted string literal
    if arg_source.startswith('"') and arg_source.endswith('"'):
        try:
            return ast.literal_eval(arg_source)
        except Exception:
            return arg_source.strip('"')

    # Fallback: return raw (maybe unresolved) expression
    return arg_source


def find_keys_in_file(path: Path) -> Set[str]:
    """Return the set of i18n keys found in a single Go file."""
    keys: Set[str] = set()
    try:
        content = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        print(f"[warn] Could not read {path} as UTF-8", file=sys.stderr)
        return keys

    module_base = path.stem  # e.g., admin.go -> admin
    module_title = module_base[:1].upper() + module_base[1:]

    for match in CALL_PATTERN.finditer(content):
        arg_src = match.group(1)
        key = extract_key(arg_src, module_title)
        if key:
            keys.add(key.lower())  # store lowercase version
    return keys


def _collect_leaf_yaml_keys(data: dict, prefix: str = "") -> Set[str]:
    """Recursively collect leaf keys from nested YAML data as dotted paths."""
    keys: Set[str] = set()
    for k, v in data.items():
        new_prefix = f"{prefix}.{k}" if prefix else k
        if isinstance(v, dict):
            keys.update(_collect_leaf_yaml_keys(v, new_prefix))
        else:
            keys.add(new_prefix)
    return keys


def main() -> None:
    # --- Collect keys used in project source ---
    if not SOURCE_DIR.exists():
        print(
            f"[error] Expected source directory not found at {SOURCE_DIR}",
            file=sys.stderr,
        )
        sys.exit(1)

    used_keys: Set[str] = set()
    for go_file in SOURCE_DIR.rglob("*.go"):
        # Skip vendor code if present
        if "vendor" in go_file.parts:
            continue
        used_keys.update(find_keys_in_file(go_file))

    # --- Collect keys defined in locale file ---
    if not LOCALE_FILE.exists():
        print(
            f"[error] Locale file not found at {LOCALE_FILE}",
            file=sys.stderr,
        )
        sys.exit(1)

    try:
        with LOCALE_FILE.open("r", encoding="utf-8") as f:
            yaml_data = yaml.safe_load(f)
    except Exception as e:
        print(f"[error] Failed to load YAML from {LOCALE_FILE}: {e}", file=sys.stderr)
        sys.exit(1)

    locale_keys = _collect_leaf_yaml_keys(yaml_data)
    locale_keys = {k.lower() for k in locale_keys}

    # --- Compute difference ---
    unused_keys = sorted(locale_keys - used_keys)

    print(f"Keys used in code: {len(used_keys)}")
    print()
    print(
        f"Unused keys in {LOCALE_FILE.relative_to(REPO_ROOT)} (count {len(unused_keys)}):\n"
    )
    for key in unused_keys:
        print(key)


if __name__ == "__main__":
    main()

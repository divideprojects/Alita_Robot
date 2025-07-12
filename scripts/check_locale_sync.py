#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
This script checks for missing i18n keys in locale files.

It compares other locale files against a source file (en.yml) to find missing keys.
It can also automatically add the missing keys from the source file to the target files,
marking them with a comment for translators.

Dependencies:
    pip install ruamel.yaml

Usage:
    python3 scripts/check_i18n_keys.py
    python3 scripts/check_i18n_keys.py --fix
"""

import argparse
import sys
from pathlib import Path
from ruamel.yaml import YAML

# --- Configuration ---
SOURCE_LANGUAGE = "en"
LOCALES_DIR = Path("locales")
TODO_COMMENT = "TODO: INCOMPLETE TRANSLATION"
# --- End Configuration ---

# Initialize a YAML parser that preserves comments and formatting
yaml = YAML()
yaml.indent(mapping=2, sequence=4, offset=2)
yaml.preserve_quotes = True


def get_all_keys(data, parent_key=""):
    """
    Recursively traverses a nested dictionary and returns a set of all key paths.
    e.g., {'a': {'b': 1}} -> {'a.b'}
    """
    keys = set()
    for k, v in data.items():
        new_key = f"{parent_key}.{k}" if parent_key else k
        if isinstance(v, dict):
            keys.update(get_all_keys(v, new_key))
        else:
            keys.add(new_key)
    return keys


def add_missing_keys_recursively(source_data, target_data, missing_keys_paths):
    """
    Recursively adds missing keys from source_data to target_data.
    It attaches a 'TODO' comment to the newly added keys.
    """
    modified = False
    for key, source_value in source_data.items():
        # Construct the full path for the current key for lookup
        current_path_tuple = tuple(key.split("."))

        # Check if the key is missing in the target
        if key not in target_data:
            # Check if this specific key or any of its children are in the missing set
            is_missing = False
            for missing_path in missing_keys_paths:
                if missing_path.startswith(key):
                    is_missing = True
                    break

            if is_missing:
                # Add the key and its value from the source
                target_data[key] = source_value
                # Add a comment to the newly added key
                comment = f" {TODO_COMMENT} "
                target_data.yaml_set_comment_before_after_key(key, before=comment)
                modified = True

        # If the key exists and both values are dicts, recurse
        elif isinstance(source_value, dict) and isinstance(target_data.get(key), dict):
            if add_missing_keys_recursively(
                source_value, target_data[key], missing_keys_paths
            ):
                modified = True

    return modified


def main():
    """Main function to run the script."""
    parser = argparse.ArgumentParser(
        description="Check for and fix missing i18n keys in YAML locale files.",
        formatter_class=argparse.RawTextHelpFormatter,
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Automatically add missing keys from the source language file (en.yml) \n"
        "to other language files, marked with a TODO comment.",
    )
    args = parser.parse_args()

    # --- 1. Load Source File ---
    source_file = LOCALES_DIR / f"{SOURCE_LANGUAGE}.yml"
    if not source_file.is_file():
        print(f"Error: Source language file not found at '{source_file}'")
        sys.exit(1)

    print(f"Source language: {source_file}")
    source_data = yaml.load(source_file)
    source_keys = get_all_keys(source_data)

    # --- 2. Process Target Files ---
    total_missing_count = 0
    files_to_fix = []

    target_files = sorted(
        f
        for f in LOCALES_DIR.glob("*.yml")
        if f.name != source_file.name and f.name != "config.yml"
    )

    for target_file in target_files:
        print(f"\n--- Checking: {target_file.name} ---")
        target_data = yaml.load(target_file)

        # Handle empty or invalid YAML files
        if not target_data:
            print("  File is empty or invalid. Skipping.")
            continue

        target_keys = get_all_keys(target_data)
        missing_keys = source_keys - target_keys

        if not missing_keys:
            print("  All keys are present. ✓")
            continue

        print(f"  Found {len(missing_keys)} missing keys:")
        for key in sorted(list(missing_keys)):
            print(f"    - {key}")

        total_missing_count += len(missing_keys)
        if args.fix:
            files_to_fix.append((target_file, target_data, missing_keys))

    # --- 3. Report Summary ---
    if total_missing_count == 0:
        print("\nAll locale files are synchronized with the source. Great work!")
        sys.exit(0)
    else:
        print(
            f"\nFound a total of {total_missing_count} missing key(s) across all files."
        )

    # --- 4. Apply Fixes if Requested ---
    if not args.fix:
        print(
            "\nTo fix these issues automatically, run the script with the --fix flag."
        )
        sys.exit(1)

    print("\nApplying fixes as requested...")
    for file_path, data, missing_keys in files_to_fix:
        print(f"  Fixing {file_path.name}...")

        # We need to reload the source data with the YAML object to preserve structure
        source_data_for_fix = yaml.load(source_file)

        # Create a deep copy for modification
        from copy import deepcopy

        modified_data = deepcopy(data)

        if add_missing_keys_recursively(
            source_data_for_fix, modified_data, missing_keys
        ):
            with open(file_path, "w", encoding="utf-8") as f:
                yaml.dump(modified_data, f)
            print(f"    -> Successfully wrote updated keys to {file_path.name}")
        else:
            print(
                f"    -> No changes were made to {file_path.name}, although missing keys were detected. This might be a bug."
            )

    print("\nFixes applied. Please review the changes and commit them.")
    sys.exit(1)  # Exit with error code to indicate that changes were made


if __name__ == "__main__":
    main()

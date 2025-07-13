#!/usr/bin/env python3
"""
Script to find keys in locale YAML files that have the same string values.
Processes all .yml files in the locales directory except config.yml.
"""

import yaml
from collections import defaultdict
import sys
from pathlib import Path
import glob


def flatten_dict(d, parent_key="", sep="."):
    """
    Flatten a nested dictionary into a single level with dot-separated keys.
    """
    items = []
    for k, v in d.items():
        new_key = f"{parent_key}{sep}{k}" if parent_key else k
        if isinstance(v, dict):
            items.extend(flatten_dict(v, new_key, sep=sep).items())
        else:
            items.append((new_key, v))
    return dict(items)


def find_duplicate_strings(yaml_file):
    """
    Find keys in YAML file that have the same string values.
    Returns a dictionary of duplicates or None if file cannot be processed.
    """
    try:
        with open(yaml_file, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
    except FileNotFoundError:
        print(f"Error: File '{yaml_file}' not found.")
        return None
    except yaml.YAMLError as e:
        print(f"Error parsing YAML file '{yaml_file}': {e}")
        return None
    except Exception as e:
        print(f"Error reading file '{yaml_file}': {e}")
        return None

    if not data:
        print(f"Warning: File '{yaml_file}' is empty or invalid.")
        return {}

    # Flatten the nested dictionary
    flat_data = flatten_dict(data)

    # Group keys by their string values
    value_to_keys = defaultdict(list)

    for key, value in flat_data.items():
        # Only consider string values (skip None, numbers, booleans, etc.)
        if isinstance(value, str) and value.strip():
            value_to_keys[value].append(key)

    # Find duplicates (values that have more than one key)
    duplicates = {value: keys for value, keys in value_to_keys.items() if len(keys) > 1}

    return duplicates


def process_locale_files(locales_dir="locales"):
    """
    Process all YAML files in the locales directory except config.yml.
    """
    locales_path = Path(locales_dir)

    if not locales_path.exists():
        print(f"Error: Directory '{locales_dir}' does not exist.")
        return

    if not locales_path.is_dir():
        print(f"Error: '{locales_dir}' is not a directory.")
        return

    # Find all .yml files in the locales directory, excluding config.yml
    yaml_files = []
    for pattern in ["*.yml", "*.yaml"]:
        yaml_files.extend(glob.glob(str(locales_path / pattern)))

    # Filter out config.yml and sort the files
    yaml_files = [
        f for f in yaml_files if not Path(f).name.lower().startswith("config.")
    ]
    yaml_files.sort()

    if not yaml_files:
        print(f"No YAML files found in '{locales_dir}' directory.")
        return

    print(f"Found {len(yaml_files)} locale files to analyze:")
    for f in yaml_files:
        print(f"  - {Path(f).name}")
    print("=" * 70)

    total_duplicates = 0

    for yaml_file in yaml_files:
        file_name = Path(yaml_file).name
        print(f"\nAnalyzing: {file_name}")
        print("-" * 50)

        duplicates = find_duplicate_strings(yaml_file)

        if duplicates is None:
            continue

        if not duplicates:
            print("No duplicate string values found.")
            continue

        file_duplicate_count = len(duplicates)
        total_duplicates += file_duplicate_count

        print(f"Found {file_duplicate_count} duplicate string values:\n")

        # Sort by number of occurrences (descending) and then by value
        sorted_duplicates = sorted(duplicates.items(), key=lambda x: (-len(x[1]), x[0]))

        for value, keys in sorted_duplicates:
            # Truncate very long strings for display
            display_value = value if len(value) <= 100 else value[:97] + "..."
            print(f"String: '{display_value}'")
            print(f"Found in {len(keys)} keys:")
            for key in sorted(keys):
                print(f"  - {key}")
            print()

    print("=" * 70)
    print(
        f"Summary: Found {total_duplicates} total duplicate string values across all locale files."
    )


def main():
    # Allow specifying a different locales directory
    locales_dir = sys.argv[1] if len(sys.argv) > 1 else "locales"

    print(f"Analyzing locale files in directory: {locales_dir}")
    print("Skipping config.yml files")
    print("=" * 70)

    process_locale_files(locales_dir)


if __name__ == "__main__":
    main()

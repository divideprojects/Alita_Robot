#!/usr/bin/env python3
"""
Script to find keys in locale YAML files that have the same string values.
Processes all .yml files in the locales directory except config.yml.
With --fix flag, moves duplicates to strings.CommonStrings with appropriate keys.
"""

import yaml
from collections import defaultdict
import sys
from pathlib import Path
import glob
import argparse
import re


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


def unflatten_dict(flat_dict, sep="."):
    """
    Convert a flattened dictionary back to nested structure.
    """
    result = {}
    for key, value in flat_dict.items():
        keys = key.split(sep)
        current = result
        for k in keys[:-1]:
            if k not in current:
                current[k] = {}
            current = current[k]
        current[keys[-1]] = value
    return result


def generate_common_key(value, existing_keys=None):
    """
    Generate a suitable key name for strings.CommonStrings based on the string value.
    """
    if existing_keys is None:
        existing_keys = set()

    # Clean the value to create a key
    # Remove HTML tags
    clean_value = re.sub(r"<[^>]+>", "", value)
    # Remove special characters and normalize
    clean_value = re.sub(r"[^\w\s]", "", clean_value)
    # Convert to lowercase and replace spaces with underscores
    clean_value = clean_value.lower().strip()
    clean_value = re.sub(r"\s+", "_", clean_value)

    # Truncate if too long
    if len(clean_value) > 50:
        clean_value = clean_value[:50]

    # Remove trailing underscores
    clean_value = clean_value.rstrip("_")

    # If empty, use generic name
    if not clean_value:
        clean_value = "common_string"

    # Ensure uniqueness
    base_key = clean_value
    counter = 1
    while clean_value in existing_keys:
        clean_value = f"{base_key}_{counter}"
        counter += 1

    return clean_value


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


def fix_duplicates(yaml_file, duplicates):
    """
    Fix duplicate strings by moving them to strings.CommonStrings and updating references.
    """
    try:
        with open(yaml_file, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)
    except Exception as e:
        print(f"Error reading file '{yaml_file}': {e}")
        return False

    if not data:
        print(f"Warning: File '{yaml_file}' is empty or invalid.")
        return False

    # Ensure strings.CommonStrings exists
    if "strings" not in data:
        data["strings"] = {}
    if "CommonStrings" not in data["strings"]:
        data["strings"]["CommonStrings"] = {}

    # Flatten the data for easier manipulation
    flat_data = flatten_dict(data)

    # Get existing CommonStrings keys to avoid conflicts
    existing_common_keys = set()
    for key in flat_data.keys():
        if key.startswith("strings.CommonStrings."):
            # Extract the key part after strings.CommonStrings.
            common_key = key[len("strings.CommonStrings.") :]
            existing_common_keys.add(common_key)

    # Process each duplicate string
    fixes_made = 0
    for value, keys in duplicates.items():
        if len(keys) < 2:
            continue

        # Generate a new key for CommonStrings
        new_common_key = generate_common_key(value, existing_common_keys)
        existing_common_keys.add(new_common_key)

        # Add to CommonStrings
        common_strings_key = f"strings.CommonStrings.{new_common_key}"
        flat_data[common_strings_key] = value

        # Remove the duplicate keys and track what was removed
        removed_keys = []
        for key in keys:
            if key in flat_data:
                del flat_data[key]
                removed_keys.append(key)

        if removed_keys:
            fixes_made += 1
            print(
                f"  Moved '{value[:50]}{'...' if len(value) > 50 else ''}' to CommonStrings.{new_common_key}"
            )
            print(f"    Removed {len(removed_keys)} duplicate keys:")
            for key in removed_keys:
                print(f"      - {key}")
            print()

    # Convert back to nested structure
    fixed_data = unflatten_dict(flat_data)

    # Save the fixed file
    try:
        with open(yaml_file, "w", encoding="utf-8") as f:
            yaml.dump(
                fixed_data,
                f,
                default_flow_style=False,
                allow_unicode=True,
                sort_keys=False,
            )
        print(f"Successfully fixed {fixes_made} duplicate strings in {yaml_file}")
        return True
    except Exception as e:
        print(f"Error writing fixed file '{yaml_file}': {e}")
        return False


def process_locale_files(locales_dir="locales", fix_mode=False):
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
    total_fixed = 0

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

        if fix_mode:
            print(f"Fixing {file_duplicate_count} duplicate string values:\n")
            if fix_duplicates(yaml_file, duplicates):
                total_fixed += file_duplicate_count
        else:
            print(f"Found {file_duplicate_count} duplicate string values:\n")

            # Sort by number of occurrences (descending) and then by value
            sorted_duplicates = sorted(
                duplicates.items(), key=lambda x: (-len(x[1]), x[0])
            )

            for value, keys in sorted_duplicates:
                # Truncate very long strings for display
                display_value = value if len(value) <= 100 else value[:97] + "..."
                print(f"String: '{display_value}'")
                print(f"Found in {len(keys)} keys:")
                for key in sorted(keys):
                    print(f"  - {key}")
                print()

    print("=" * 70)
    if fix_mode:
        print(
            f"Summary: Fixed {total_fixed} duplicate string values across all locale files."
        )
        if total_fixed > 0:
            print(
                "\nIMPORTANT: You'll need to update your Go code to reference the new CommonStrings keys."
            )
            print(
                "The old keys have been removed and replaced with references to strings.CommonStrings.*"
            )
    else:
        print(
            f"Summary: Found {total_duplicates} total duplicate string values across all locale files."
        )
        if total_duplicates > 0:
            print("\nTo fix these duplicates, run the script with the --fix flag:")
            print(f"  python {sys.argv[0]} --fix")


def main():
    parser = argparse.ArgumentParser(
        description="Find and optionally fix duplicate strings in locale YAML files"
    )
    parser.add_argument(
        "locales_dir",
        nargs="?",
        default="locales",
        help="Directory containing locale files (default: locales)",
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Fix duplicates by moving them to strings.CommonStrings",
    )

    args = parser.parse_args()

    print(f"Analyzing locale files in directory: {args.locales_dir}")
    print("Skipping config.yml files")
    if args.fix:
        print("FIX MODE: Will move duplicates to strings.CommonStrings")
    print("=" * 70)

    process_locale_files(args.locales_dir, args.fix)


if __name__ == "__main__":
    main()

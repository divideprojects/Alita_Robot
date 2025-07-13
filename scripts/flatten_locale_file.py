import yaml
import sys
from collections.abc import Mapping


def flatten_yaml(input_file, output_file):
    """
    Reads a nested YAML file, flattens its keys, converts them to lowercase,
    and writes the result to a new file.

    This script specifically targets a structure where all content to be flattened
    is under a top-level 'strings' key. It converts boolean values to capitalized
    string versions of their original keys (e.g., {'yes': True} -> 'Yes').

    Args:
        input_file (str): The path to the input nested YAML file.
        output_file (str): The path where the flattened YAML file will be saved.
    """
    try:
        with open(input_file, "r") as f:
            data = yaml.safe_load(f)
    except FileNotFoundError:
        print(f"Error: Input file not found at '{input_file}'")
        sys.exit(1)
    except yaml.YAMLError as e:
        print(f"Error parsing YAML file: {e}")
        sys.exit(1)

    if "strings" not in data:
        print("Error: Input YAML must contain a top-level 'strings' key.")
        sys.exit(1)

    nested_dict = data["strings"]
    flattened_dict = {}

    def flatten(d, parent_key=""):
        """
        Recursively flattens a dictionary, converting all keys to lowercase.
        """
        for k, v in d.items():
            # Convert the current key part to lowercase for the new flattened key.
            new_key_part = str(k).lower()
            new_key = f"{parent_key}.{new_key_part}" if parent_key else new_key_part

            if isinstance(v, Mapping):
                flatten(v, new_key)
            else:
                # As per the request, convert boolean values to capitalized strings
                # based on their original, un-lowercased key.
                if isinstance(v, bool):
                    value_from_key = str(k).capitalize()
                    flattened_dict[new_key] = value_from_key
                else:
                    flattened_dict[new_key] = v

    flatten(nested_dict)

    # Prepare the final structure for output
    output_data = {"strings": flattened_dict}

    try:
        with open(output_file, "w") as f:
            # Dumper settings for clean, readable output
            yaml.dump(
                output_data, f, default_flow_style=False, sort_keys=False, indent=2
            )
        print(f"Successfully flattened '{input_file}' and saved to '{output_file}'")
    except IOError as e:
        print(f"Error writing to output file: {e}")
        sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python flatten_yaml.py <input_file.yml> <output_file.yml>")
        sys.exit(1)

    input_path = sys.argv[1]
    output_path = sys.argv[2]
    flatten_yaml(input_path, output_path)

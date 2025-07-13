#!/usr/bin/env python3
"""
Script to check for hardcoded strings in Alita Robot bot codebase.
This script identifies strings that should be using i18n keys instead of being hardcoded.

Usage examples:
    # Check all hardcoded strings with statistics
    python scripts/check_hardcoded_strings.py

    # Show only high-severity (user-facing) strings
    python scripts/check_hardcoded_strings.py --severity high

    # Show detailed information for all strings
    python scripts/check_hardcoded_strings.py --all

    # Export results to JSON for further processing
    python scripts/check_hardcoded_strings.py --export report.json

    # Show only statistics
    python scripts/check_hardcoded_strings.py --stats-only

    # Scan a different directory
    python scripts/check_hardcoded_strings.py --directory alita/modules
"""

import os
import re
import argparse
import json
from typing import List, Dict
from dataclasses import dataclass


@dataclass
class HardcodedString:
    file_path: str
    line_number: int
    line_content: str
    string_content: str
    pattern_type: str
    severity: str = "medium"


class HardcodedStringChecker:
    def __init__(self, root_dir: str = "alita"):
        self.root_dir = root_dir
        self.hardcoded_strings: List[HardcodedString] = []

        # Patterns to look for hardcoded strings
        self.patterns = {
            # Message replies with hardcoded strings
            "msg_reply": {
                "pattern": r'msg\.Reply\([^,]*,\s*"([^"]+)"',
                "description": "msg.Reply() with hardcoded string",
                "severity": "high",
            },
            "msg_reply_text": {
                "pattern": r'msg\.ReplyText\([^,]*,\s*"([^"]+)"',
                "description": "msg.ReplyText() with hardcoded string",
                "severity": "high",
            },
            # SendMessage with hardcoded strings
            "send_message": {
                "pattern": r'SendMessage\([^,]*,\s*"([^"]+)"',
                "description": "SendMessage() with hardcoded string",
                "severity": "high",
            },
            "send_msg": {
                "pattern": r'SendMsg\([^,]*,\s*"([^"]+)"',
                "description": "SendMsg() with hardcoded string",
                "severity": "high",
            },
            # Format strings
            "fmt_sprintf": {
                "pattern": r'fmt\.Sprintf\(\s*"([^"]+)"',
                "description": "fmt.Sprintf() with hardcoded format string",
                "severity": "medium",
            },
            "fmt_errorf": {
                "pattern": r'fmt\.Errorf\(\s*"([^"]+)"',
                "description": "fmt.Errorf() with hardcoded error message",
                "severity": "medium",
            },
            # Log messages (lower priority)
            "log_error": {
                "pattern": r'log\.Error\([^,]*,\s*"([^"]+)"',
                "description": "log.Error() with hardcoded message",
                "severity": "low",
            },
            "log_info": {
                "pattern": r'log\.Info\([^,]*,\s*"([^"]+)"',
                "description": "log.Info() with hardcoded message",
                "severity": "low",
            },
            # Return statements with user-facing strings
            "return_error": {
                "pattern": r'return\s+.*errors\.New\(\s*"([^"]+)"\s*\)',
                "description": "Return error with hardcoded message",
                "severity": "medium",
            },
            # Bot API calls with hardcoded text
            "answer_callback": {
                "pattern": r'AnswerCallbackQuery\([^,]*,\s*"([^"]+)"',
                "description": "AnswerCallbackQuery() with hardcoded text",
                "severity": "high",
            },
            # Edit message text
            "edit_message": {
                "pattern": r'EditMessageText\([^,]*,\s*"([^"]+)"',
                "description": "EditMessageText() with hardcoded text",
                "severity": "high",
            },
        }

        # Strings to ignore (technical strings, not user-facing)
        self.ignore_patterns = [
            r"^[A-Z_]+$",  # Constants in ALL_CAPS
            r"^[a-z_]+$",  # Simple identifiers
            r"^\d+$",  # Numbers only
            r"^https?://",  # URLs
            r"^/[a-zA-Z_]+",  # Command patterns
            r"^[<>{}[\]()]+$",  # Markup/syntax characters
            r"^\s*$",  # Whitespace only
            r"^[.,!?;:]+$",  # Punctuation only
            r"^%[a-zA-Z]",  # Format specifiers
            r"^\$\{.*\}$",  # Template variables
            r"^[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+$",  # Package/module names
        ]

        # File extensions to check
        self.file_extensions = [".go"]

        # Directories to skip
        self.skip_dirs = {"vendor", "testdata", ".git", "node_modules"}

    def should_ignore_string(self, string_content: str) -> bool:
        """Check if a string should be ignored (not user-facing)."""
        if len(string_content.strip()) < 3:
            return True

        for pattern in self.ignore_patterns:
            if re.match(pattern, string_content.strip()):
                return True

        return False

    def is_user_facing_string(self, string_content: str, context: str) -> bool:
        """Determine if a string is likely user-facing."""
        # Check for common user-facing indicators
        user_facing_indicators = [
            r"\b(you|your|user|admin|permission|error|success|warning|info)\b",
            r"\b(please|sorry|cannot|can\'t|don\'t|won\'t)\b",
            r"\b(chat|group|channel|message|command)\b",
            r"[.!?]$",  # Ends with sentence punctuation
        ]

        string_lower = string_content.lower()
        for indicator in user_facing_indicators:
            if re.search(indicator, string_lower):
                return True

        # Check context for user-facing function calls
        user_facing_contexts = [
            "Reply",
            "SendMessage",
            "SendMsg",
            "AnswerCallback",
            "EditMessage",
        ]

        for ctx in user_facing_contexts:
            if ctx in context:
                return True

        return False

    def check_file(self, file_path: str) -> None:
        """Check a single file for hardcoded strings."""
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                lines = f.readlines()

            for line_num, line in enumerate(lines, 1):
                line_stripped = line.strip()

                # Skip comments and empty lines
                if line_stripped.startswith("//") or not line_stripped:
                    continue

                # Skip lines that already use tr.GetString
                if "tr.GetString" in line or "strings." in line:
                    continue

                # Check each pattern
                for pattern_name, pattern_info in self.patterns.items():
                    matches = re.finditer(pattern_info["pattern"], line)

                    for match in matches:
                        string_content = match.group(1)

                        # Skip if should be ignored
                        if self.should_ignore_string(string_content):
                            continue

                        # For medium/low severity, check if it's user-facing
                        if pattern_info["severity"] in ["medium", "low"]:
                            if not self.is_user_facing_string(string_content, line):
                                continue

                        hardcoded_string = HardcodedString(
                            file_path=file_path,
                            line_number=line_num,
                            line_content=line.strip(),
                            string_content=string_content,
                            pattern_type=pattern_info["description"],
                            severity=pattern_info.get("severity", "medium"),
                        )

                        self.hardcoded_strings.append(hardcoded_string)

        except Exception as e:
            print(f"Error reading file {file_path}: {e}")

    def scan_directory(self) -> None:
        """Scan the directory for hardcoded strings."""
        for root, dirs, files in os.walk(self.root_dir):
            # Skip certain directories
            dirs[:] = [d for d in dirs if d not in self.skip_dirs]

            for file in files:
                if any(file.endswith(ext) for ext in self.file_extensions):
                    file_path = os.path.join(root, file)
                    self.check_file(file_path)

    def get_statistics(self) -> Dict[str, int]:
        """Get statistics about found hardcoded strings."""
        stats = {
            "total": len(self.hardcoded_strings),
            "high": 0,
            "medium": 0,
            "low": 0,
            "files_affected": len(set(hs.file_path for hs in self.hardcoded_strings)),
        }

        for hs in self.hardcoded_strings:
            stats[hs.severity] += 1

        return stats

    def print_results(
        self, show_all: bool = False, severity_filter: str | None = None
    ) -> None:
        """Print the results of the scan."""
        if not self.hardcoded_strings:
            print("✅ No hardcoded strings found!")
            return

        # Filter by severity if specified
        strings_to_show = self.hardcoded_strings
        if severity_filter:
            strings_to_show = [
                hs for hs in self.hardcoded_strings if hs.severity == severity_filter
            ]

        # Group by file
        by_file = {}
        for hs in strings_to_show:
            if hs.file_path not in by_file:
                by_file[hs.file_path] = []
            by_file[hs.file_path].append(hs)

        print(f"\n🔍 Found {len(strings_to_show)} hardcoded strings:")
        print("=" * 60)

        for file_path, strings in sorted(by_file.items()):
            print(f"\n📁 {file_path}")
            print("-" * 40)

            for hs in sorted(strings, key=lambda x: x.line_number):
                severity_emoji = {"high": "🔴", "medium": "🟡", "low": "🟢"}
                emoji = severity_emoji.get(hs.severity, "⚪")

                if show_all or hs.severity == "high":
                    print(f"  {emoji} Line {hs.line_number}: {hs.pattern_type}")
                    print(f'     String: "{hs.string_content}"')
                    print(f"     Context: {hs.line_content}")
                    print()
                else:
                    print(
                        f"  {emoji} Line {hs.line_number}: \"{hs.string_content[:50]}{'...' if len(hs.string_content) > 50 else ''}\""
                    )

    def export_json(self, output_file: str) -> None:
        """Export results to JSON file."""
        data = {
            "statistics": self.get_statistics(),
            "hardcoded_strings": [
                {
                    "file_path": hs.file_path,
                    "line_number": hs.line_number,
                    "line_content": hs.line_content,
                    "string_content": hs.string_content,
                    "pattern_type": hs.pattern_type,
                    "severity": hs.severity,
                }
                for hs in self.hardcoded_strings
            ],
        }

        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)

        print(f"📄 Results exported to {output_file}")


def main():
    parser = argparse.ArgumentParser(
        description="Check for hardcoded strings in Alita Robot bot"
    )
    parser.add_argument(
        "--directory", "-d", default="alita", help="Directory to scan (default: alita)"
    )
    parser.add_argument(
        "--all", "-a", action="store_true", help="Show all details for each string"
    )
    parser.add_argument(
        "--severity",
        "-s",
        choices=["high", "medium", "low"],
        help="Filter by severity level",
    )
    parser.add_argument("--export", "-e", help="Export results to JSON file")
    parser.add_argument(
        "--stats-only", action="store_true", help="Show only statistics"
    )

    args = parser.parse_args()

    if not os.path.exists(args.directory):
        print(f"❌ Directory '{args.directory}' does not exist!")
        return 1

    print(f"🔍 Scanning directory: {args.directory}")
    print("Looking for hardcoded strings that should use i18n...")

    checker = HardcodedStringChecker(args.directory)
    checker.scan_directory()

    # Print statistics
    stats = checker.get_statistics()
    print(f"\n📊 Statistics:")
    print(f"   Total hardcoded strings: {stats['total']}")
    print(f"   High severity (user-facing): {stats['high']}")
    print(f"   Medium severity: {stats['medium']}")
    print(f"   Low severity: {stats['low']}")
    print(f"   Files affected: {stats['files_affected']}")

    if not args.stats_only:
        checker.print_results(show_all=args.all, severity_filter=args.severity)

    if args.export:
        checker.export_json(args.export)

    # Suggest next steps
    if stats["high"] > 0:
        print(f"\n💡 Recommendations:")
        print(
            f"   1. Focus on {stats['high']} high-severity strings first (user-facing)"
        )
        print(f"   2. Replace with tr.GetString() or tr.GetStringWithError() calls")
        print(f"   3. Add corresponding keys to locales/en.yml")
        print(f"   4. Run this script again to track progress")

    return 0 if stats["total"] == 0 else 1


if __name__ == "__main__":
    exit(main())

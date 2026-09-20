#!/usr/bin/env python3
"""
Ops Phoenix - Self-Sufficient Ops Agent

A comprehensive ops agent that asks for all required info before executing.
Works standalone on VM or with Claude Code.

Usage:
    python3 ops-phoenix.py                    # Interactive mode
    python3 ops-phoenix.py --action detect    # Specific action
    python3 ops-phoenix.py --setup           # First-time setup
"""

import json
import os
import sys
import subprocess
import re
from datetime import datetime
from pathlib import Path

# Paths
HOME_DIR = Path.home()
CONFIG_DIR = HOME_DIR / ".ops-phoenix"
CONFIG_FILE = CONFIG_DIR / "config.json"
HISTORY_FILE = CONFIG_DIR / "history.log"

# Defaults
DEFAULT_GRAFANA_TOKEN_PATH = HOME_DIR / "Downloads/Archive/keys/grafana-api-token"
DEFAULT_GITHUB_REPO = "muthuishere/reqsume"
DEFAULT_TIME_WINDOW = "1h"
DEFAULT_API_URL = "https://api.opusmax.pro"

class OpsPhoenix:
    def __init__(self):
        self.config = self.load_config()
        self.action = None
        self.time_window = None
        self.environment = None
        self.api_key = None
        self.grafana_token_path = None
        self.github_repo = None

    def load_config(self):
        """Load config from file"""
        if CONFIG_FILE.exists():
            with open(CONFIG_FILE) as f:
                return json.load(f)
        return {}

    def save_config(self):
        """Save config to file"""
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        with open(CONFIG_FILE, 'w') as f:
            json.dump(self.config, f, indent=2)

    def log(self, message):
        """Log message to console and file"""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        line = f"[{timestamp}] {message}"
        print(line)

        # Append to history
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        with open(HISTORY_FILE, 'a') as f:
            f.write(line + "\n")

    def ask(self, question, default=None, options=None, required=True):
        """Ask user a question and return answer"""
        if options:
            print(f"\n{question}")
            for i, opt in enumerate(options, 1):
                default_marker = " (default)" if opt == default else ""
                print(f"  {i}. {opt}{default_marker}")
            while True:
                try:
                    choice = input(f"Enter choice (1-{len(options)}): ").strip()
                    if not choice and default:
                        return default
                    idx = int(choice) - 1
                    if 0 <= idx < len(options):
                        return options[idx]
                    print(f"Please enter 1-{len(options)}")
                except ValueError:
                    print("Please enter a number")
        else:
            prompt = f"\n{question}"
            if default:
                prompt += f" [{default}]"
            if required:
                prompt += " *"
            prompt += ": "

            while True:
                answer = input(prompt).strip()
                if answer:
                    return answer
                elif default:
                    return default
                elif not required:
                    return ""
                print("This field is required")

    def ask_password(self, question):
        """Ask for sensitive input (password/API key)"""
        import getpass
        prompt = f"\n{question} (hidden): "
        while True:
            answer = getpass.getpass(prompt).strip()
            if answer:
                return answer
            print("This field is required")

    def setup(self):
        """First-time setup - ask for all required info"""
        self.log("=== OPS PHOENIX SETUP ===")
        print("\nWelcome to Ops Phoenix! I'll help you set up the agent.")
        print("You can press Enter to accept defaults.\n")

        # Ask for each required item
        self.log("Collecting configuration...")

        # Time window
        self.time_window = self.ask(
            "Time window for error checks?",
            default=DEFAULT_TIME_WINDOW,
            options=["1h", "6h", "24h", "7d"],
            required=False
        )

        # Environment
        self.environment = self.ask(
            "Environment?",
            default="production",
            options=["production", "dev"],
            required=False
        )

        # GitHub repo
        self.github_repo = self.ask(
            "GitHub repository?",
            default=DEFAULT_GITHUB_REPO,
            required=False
        )

        # Grafana token path
        default_grafana = str(DEFAULT_GRAFANA_TOKEN_PATH)
        current_default = self.config.get('grafana_token_path', default_grafana)

        if Path(current_default).exists():
            use_default = input(f"\nUse default Grafana token path? [{current_default}] (y/n): ").strip().lower()
            if use_default == 'y' or not use_default:
                self.grafana_token_path = current_default
            else:
                self.grafana_token_path = self.ask("Enter Grafana token path:")
        else:
            self.grafana_token_path = self.ask("Enter Grafana token path:", required=True)

        # API key
        if os.environ.get('ANTHROPIC_API_KEY'):
            use_env = input("\nUse ANTHROPIC_API_KEY from environment? (y/n): ").strip().lower()
            if use_env == 'y' or not use_env:
                self.api_key = os.environ.get('ANTHROPIC_API_KEY')
                self.log("Using API key from environment")
            else:
                self.api_key = self.ask_password("Enter Claude API key")
        else:
            self.api_key = self.ask_password("Enter Claude API key (or set ANTHROPIC_API_KEY env var)")

        # Save config
        self.config = {
            'time_window': self.time_window,
            'environment': self.environment,
            'github_repo': self.github_repo,
            'grafana_token_path': self.grafana_token_path,
            'setup_complete': True,
            'setup_date': datetime.now().isoformat()
        }
        self.save_config()

        self.log("Setup complete! Config saved to: " + str(CONFIG_FILE))

        # Verify
        print("\n=== CONFIGURATION SUMMARY ===")
        for key, value in self.config.items():
            if 'token' not in key.lower() and 'key' not in key.lower():
                self.log(f"  {key}: {value}")

        # Verify GitHub auth
        print("\nVerifying GitHub authentication...")
        result = subprocess.run(['gh', 'auth', 'status'], capture_output=True, text=True)
        if result.returncode == 0:
            self.log("GitHub: Authenticated")
        else:
            print("Warning: GitHub CLI not authenticated. Run 'gh auth login' to fix.")

        # Test Grafana token
        print("\nVerifying Grafana token...")
        if Path(self.grafana_token_path).exists():
            self.log("Grafana token: Found")
        else:
            print("Warning: Grafana token not found at: " + self.grafana_token_path)

        print("\n=== SETUP COMPLETE ===")
        print("\nNext steps:")
        print("  1. Run: python3 ops-phoenix.py --action detect")
        print("  2. Or add to crontab for automated runs")

    def status(self):
        """Show current configuration and status"""
        self.log("=== OPS PHOENIX STATUS ===")

        # Load fresh config
        self.config = self.load_config()

        if not self.config.get('setup_complete'):
            print("\nNot configured yet. Run: python3 ops-phoenix.py --setup")
            return

        print("\nConfiguration:")
        for key, value in self.config.items():
            if 'token' in key.lower() or 'key' in key.lower():
                value = "(set)" if value else "(not set)"
            self.log(f"  {key}: {value}")

        print("\nEnvironment variables:")
        api_key = os.environ.get('ANTHROPIC_API_KEY')
        self.log(f"  ANTHROPIC_API_KEY: {'(set)' if api_key else '(not set)'}")
        self.log(f"  ANTHROPIC_BASE_URL: {os.environ.get('ANTHROPIC_BASE_URL', DEFAULT_API_URL)}")

        print("\nRecent history:")
        if HISTORY_FILE.exists():
            with open(HISTORY_FILE) as f:
                lines = f.readlines()
                for line in lines[-5:]:
                    print(f"  {line.strip()}")

    def load_from_config(self):
        """Load values from config file"""
        self.time_window = self.config.get('time_window', DEFAULT_TIME_WINDOW)
        self.environment = self.config.get('environment', 'production')
        self.github_repo = self.config.get('github_repo', DEFAULT_GITHUB_REPO)
        self.grafana_token_path = self.config.get('grafana_token_path', str(DEFAULT_GRAFANA_TOKEN_PATH))

        # API key from env or config
        self.api_key = os.environ.get('ANTHROPIC_API_KEY')
        if not self.api_key:
            self.api_key = self.config.get('api_key')

    def detect(self):
        """Detect errors in logs"""
        self.log("=== OPS PHOENIX: DETECT ===")
        self.log(f"Environment: {self.environment}")
        self.log(f"Time window: {self.time_window}")

        # Load config if not already
        if not self.time_window:
            self.load_from_config()

        # Check prerequisites
        if not self.api_key:
            self.log("ERROR: No API key. Run --setup or set ANTHROPIC_API_KEY")
            return

        if not Path(self.grafana_token_path).exists():
            self.log(f"ERROR: Grafana token not found: {self.grafana_token_path}")
            return

        # Query Loki
        self.log("Querying Loki for errors...")
        errors = self.query_loki()

        if errors:
            self.log(f"Found {len(errors)} errors:")
            for err in errors[:5]:
                self.log(f"  - {err}")
        else:
            self.log("No errors found in reqsume-app containers")

        return errors

    def query_loki(self):
        """Query Loki for errors"""
        with open(self.grafana_token_path) as f:
            token = f.read().strip()

        # Time range
        ranges = {
            "1h": ("now-1h", "now"),
            "6h": ("now-6h", "now"),
            "24h": ("now-24h", "now"),
            "7d": ("now-7d", "now"),
        }
        from_ts, to_ts = ranges.get(self.time_window, ("now-1h", "now"))

        query = {
            "queries": [{
                "expr": '{container=~"reqsume-app-web.*|reqsume-api.*", host=~"app1|app2"} | json | level=~"(?i)error|fatal"',
                "refId": "A",
                "datasource": {"type": "loki", "uid": "loki"},
                "maxLines": 100
            }],
            "from": from_ts,
            "to": to_ts
        }

        result = subprocess.run([
            "curl", "-sS",
            "-H", f"Authorization: Bearer {token}",
            "-H", "Content-Type: application/json",
            "-X", "POST",
            "https://observability.deemwar.com/api/ds/query?ds_type=loki",
            "--data", json.dumps(query)
        ], capture_output=True, text=True)

        if result.returncode != 0:
            self.log(f"ERROR: Loki query failed: {result.stderr}")
            return []

        try:
            data = json.loads(result.stdout)
            frames = data.get("results", {}).get("A", {}).get("frames", [])
            if not frames:
                return []

            vals = frames[0].get("data", {}).get("values", [])
            if len(vals) <= 2:
                return []

            return [l for l in vals[2] if l and l.strip()]
        except json.JSONDecodeError as e:
            self.log(f"ERROR: Failed to parse Loki response: {e}")
            return []

    def run(self):
        """Main run method"""
        self.log("=== OPS PHOENIX ===")

        # Load config
        self.load_from_config()

        # Execute based on action
        if self.action == "setup":
            self.setup()
        elif self.action == "status":
            self.status()
        elif self.action == "detect":
            self.detect()
        elif self.action in ["analyze", "issue", "fix", "deploy", "full"]:
            errors = self.detect()
            if errors:
                self.log(f"Found {len(errors)} errors - analysis not yet implemented")
                self.log("Use --action detect to see errors")
            else:
                self.log("No errors to analyze")
        else:
            # Interactive mode
            if not self.config.get('setup_complete'):
                print("\nOps Phoenix not configured yet!")
                setup_now = input("Run setup now? (y/n): ").strip().lower()
                if setup_now == 'y':
                    self.setup()
                    return

            print("\n=== OPS PHOENIX ===")
            action = self.ask(
                "What action do you want to perform?",
                default="detect",
                options=["detect", "status", "setup"]
            )

            if action == "status":
                self.status()
            elif action == "setup":
                self.setup()
            else:
                self.detect()


def main():
    agent = OpsPhoenix()

    # Parse arguments
    args = sys.argv[1:]
    i = 0
    while i < len(args):
        arg = args[i]
        if arg in ["--action", "-a"]:
            if i + 1 < len(args):
                agent.action = args[i + 1]
                i += 2
            else:
                print("Error: --action requires a value")
                sys.exit(1)
        elif arg == "--setup":
            agent.action = "setup"
            i += 1
        elif arg == "--status":
            agent.action = "status"
            i += 1
        elif arg == "--window" and i + 1 < len(args):
            agent.time_window = args[i + 1]
            i += 2
        elif arg == "--help" or arg == "-h":
            print(__doc__)
            sys.exit(0)
        else:
            print(f"Unknown argument: {arg}")
            print("Use --help for usage")
            sys.exit(1)

    # Run
    agent.run()


if __name__ == "__main__":
    main()
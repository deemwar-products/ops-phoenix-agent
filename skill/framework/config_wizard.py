#!/usr/bin/env python3
"""
Ops Phoenix - Configuration Wizard

Asks user questions to configure the framework for their project.
Questions are asked at startup if config is incomplete.

Usage:
    python3 config_wizard.py              # Interactive mode
    python3 config_wizard.py --config     # Show current config
    python3 config_wizard.py --reset      # Reset config
"""

import json
import os
import sys
from pathlib import Path

# Config file location
CONFIG_DIR = Path.home() / ".ops-phoenix"
CONFIG_FILE = CONFIG_DIR / "config.json"
HISTORY_FILE = CONFIG_DIR / "history.log"

# Default values
DEFAULTS = {
    "observability": {
        "type": "grafana_cloud",
        "grafana_url": "",
        "grafana_token_path": "~/Downloads/Archive/keys/grafana-api-token",
        "loki_url": "",
        "loki_user": "",
        "loki_password": "",
        "container_patterns": ["app-*", "api-*"],
        "host_patterns": ["prod-*"],
        "error_patterns": ['level=~"(?i)error|fatal"']
    },
    "github": {
        "repo": "",
        "token_env_var": "GITHUB_TOKEN",
        "base_branch": "main",
        "default_labels": ["sre-alert"]
    },
    "cicd": {
        "type": "github_actions",
        "workflow_name": "Deploy to Production"
    },
    "deployment": {
        "auto_merge": True,
        "monitor_timeout_seconds": 300,
        "max_retries": 2
    }
}

class ConfigWizard:
    def __init__(self):
        self.config = self.load_config()
        self.changes = False

    def log(self, message, style=""):
        """Print styled message"""
        if style == "header":
            print(f"\n{'='*60}")
            print(f"  {message}")
            print(f"{'='*60}")
        elif style == "section":
            print(f"\n--- {message} ---")
        elif style == "success":
            print(f"✓ {message}")
        elif style == "error":
            print(f"✗ {message}")
        elif style == "info":
            print(f"  {message}")
        else:
            print(message)

    def ask(self, question, default=None, required=True, options=None):
        """Ask user a question"""
        if options:
            print(f"\n{question}")
            for i, opt in enumerate(options, 1):
                marker = " (default)" if opt == default else ""
                print(f"  {i}. {opt}{marker}")
            while True:
                try:
                    choice = input(f"Enter choice [1-{len(options)}]: ").strip()
                    if not choice and default:
                        idx = options.index(default)
                        return options[idx]
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
                if not required:
                    return ""
                if default:
                    return default
                print("This field is required")

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
        self.log("Configuration saved!", "success")

    def show_current_config(self):
        """Display current configuration"""
        self.log("CURRENT CONFIGURATION", "header")

        if not self.config:
            print("No configuration found. Run wizard to set up.")
            return

        print(json.dumps(self.config, indent=2))

    def reset_config(self):
        """Reset configuration"""
        confirm = input("Are you sure you want to reset all configuration? (yes/no): ").strip().lower()
        if confirm == "yes":
            if CONFIG_FILE.exists():
                CONFIG_FILE.unlink()
            if HISTORY_FILE.exists():
                HISTORY_FILE.unlink()
            self.config = {}
            self.log("Configuration reset!", "success")
        else:
            self.log("Reset cancelled")

    def run_wizard(self):
        """Run the configuration wizard"""
        self.log("OPS PHOENIX CONFIGURATION WIZARD", "header")
        print("""
This wizard will help you configure Ops Phoenix for your project.
Press Enter to accept defaults shown in brackets.

""")

        # Check if we need to configure or can skip
        if self.config and input("Configuration exists. Reconfigure? (yes/no): ").strip().lower() != 'yes':
            self.log("Keeping existing configuration", "info")
            return

        # ========================================
        # STEP 1: Observability Provider
        # ========================================
        self.log("OBSERVABILITY SETUP", "section")

        obs_type = self.ask(
            "What logging/monitoring system do you use?",
            default=DEFAULTS["observability"]["type"],
            options=[
                "grafana_cloud",
                "loki_direct",
                "grafana_self_hosted"
            ]
        )

        self.config.setdefault("observability", {})
        self.config["observability"]["type"] = obs_type

        if obs_type == "grafana_cloud":
            self.config["observability"]["grafana_url"] = self.ask(
                "Grafana Cloud URL (e.g., https://grafana.company.com)",
                default=self.config["observability"].get("grafana_url", ""),
                required=True
            )
            self.config["observability"]["grafana_token_path"] = self.ask(
                "Path to Grafana API token file",
                default=self.config["observability"].get("grafana_token_path", DEFAULTS["observability"]["grafana_token_path"]),
                required=False
            )
            self.config["observability"]["loki_url"] = ""
            self.config["observability"]["loki_user"] = ""
            self.config["observability"]["loki_password"] = ""

        elif obs_type == "loki_direct":
            self.config["observability"]["loki_url"] = self.ask(
                "Loki URL (e.g., http://loki:3100)",
                default=self.config["observability"].get("loki_url", "http://localhost:3100"),
                required=True
            )
            self.config["observability"]["loki_user"] = self.ask(
                "Loki username (or press Enter for none)",
                default="",
                required=False
            )
            self.config["observability"]["loki_password"] = self.ask(
                "Loki password (or press Enter for none)",
                default="",
                required=False
            )
            self.config["observability"]["grafana_url"] = ""
            self.config["observability"]["grafana_token_path"] = ""

        elif obs_type == "grafana_self_hosted":
            self.config["observability"]["grafana_url"] = self.ask(
                "Self-hosted Grafana URL (e.g., http://grafana:3000)",
                default=self.config["observability"].get("grafana_url", ""),
                required=True
            )
            self.config["observability"]["grafana_token_path"] = self.ask(
                "Path to Grafana API token file",
                default=self.config["observability"].get("grafana_token_path", ""),
                required=True
            )

        # Container filters
        self.log("Container Filters", "section")
        print("Configure which containers/services to monitor:")

        containers_str = self.ask(
            "Container name patterns (regex, comma-separated)",
            default=",".join(DEFAULTS["observability"]["container_patterns"]),
            required=False
        )
        self.config["observability"]["container_patterns"] = [
            c.strip() for c in containers_str.split(",") if c.strip()
        ]

        hosts_str = self.ask(
            "Host/VM name patterns (regex, comma-separated)",
            default=",".join(DEFAULTS["observability"]["host_patterns"]),
            required=False
        )
        self.config["observability"]["host_patterns"] = [
            h.strip() for h in hosts_str.split(",") if h.strip()
        ]

        errors_str = self.ask(
            "Error patterns to detect (Loki regex, comma-separated)",
            default=",".join(DEFAULTS["observability"]["error_patterns"]),
            required=False
        )
        self.config["observability"]["error_patterns"] = [
            e.strip() for e in errors_str.split(",") if e.strip()
        ]

        # ========================================
        # STEP 2: GitHub Configuration
        # ========================================
        self.log("GITHUB SETUP", "section")

        repo = self.ask(
            "GitHub repository (owner/repo format, e.g., myorg/myapp)",
            default=self.config.get("github", {}).get("repo", ""),
            required=True
        )
        self.config.setdefault("github", {})
        self.config["github"]["repo"] = repo

        self.config["github"]["base_branch"] = self.ask(
            "Default branch to create PRs against",
            default=self.config["github"].get("base_branch", "main"),
            required=False
        )

        self.config["github"]["token_env_var"] = self.ask(
            "Environment variable name for GitHub token",
            default=self.config["github"].get("token_env_var", "GITHUB_TOKEN"),
            required=False
        )

        labels_str = self.ask(
            "Labels to add to issues/PRs (comma-separated)",
            default=",".join(DEFAULTS["github"]["default_labels"]),
            required=False
        )
        self.config["github"]["default_labels"] = [
            l.strip() for l in labels_str.split(",") if l.strip()
        ]

        # ========================================
        # STEP 3: CI/CD Configuration
        # ========================================
        self.log("CI/CD SETUP", "section")

        cicd_type = self.ask(
            "Which CI/CD system triggers deployments?",
            default=DEFAULTS["cicd"]["type"],
            options=[
                "github_actions",
                "gitlab_ci",
                "jenkins",
                "manual"
            ]
        )
        self.config.setdefault("cicd", {})
        self.config["cicd"]["type"] = cicd_type

        if cicd_type == "github_actions":
            self.config["cicd"]["workflow_name"] = self.ask(
                "Workflow name to trigger for deployment",
                default=self.config["cicd"].get("workflow_name", "Deploy to Production"),
                required=False
            )

        elif cicd_type == "gitlab_ci":
            self.config["cicd"]["pipeline_job"] = self.ask(
                "GitLab CI job name to trigger",
                default=self.config["cicd"].get("pipeline_job", "deploy-production"),
                required=False
            )

        elif cicd_type == "jenkins":
            self.config["cicd"]["job_name"] = self.ask(
                "Jenkins job name",
                default=self.config["cicd"].get("job_name", ""),
                required=True
            )
            self.config["cicd"]["jenkins_url"] = self.ask(
                "Jenkins URL",
                default=self.config["cicd"].get("jenkins_url", "http://jenkins:8080"),
                required=True
            )

        # ========================================
        # STEP 4: Deployment Settings
        # ========================================
        self.log("DEPLOYMENT SETTINGS", "section")

        auto_merge = self.ask(
            "Auto-merge PRs and deploy?",
            default="yes" if DEFAULTS["deployment"]["auto_merge"] else "no",
            options=["yes", "no"]
        )
        self.config.setdefault("deployment", {})
        self.config["deployment"]["auto_merge"] = (auto_merge == "yes")

        timeout = self.ask(
            "Workflow monitoring timeout (seconds)",
            default=str(DEFAULTS["deployment"]["monitor_timeout_seconds"]),
            required=False
        )
        self.config["deployment"]["monitor_timeout_seconds"] = int(timeout)

        retries = self.ask(
            "Max retries on deploy failure",
            default=str(DEFAULTS["deployment"]["max_retries"]),
            required=False
        )
        self.config["deployment"]["max_retries"] = int(retries)

        # ========================================
        # STEP 5: Summary
        # ========================================
        self.log("CONFIGURATION SUMMARY", "header")

        print(f"""
Observability:
  Type: {self.config['observability']['type']}
  Containers: {', '.join(self.config['observability']['container_patterns'])}
  Error patterns: {len(self.config['observability']['error_patterns'])} configured

GitHub:
  Repository: {self.config['github']['repo']}
  Base branch: {self.config['github']['base_branch']}
  Labels: {', '.join(self.config['github']['default_labels'])}

CI/CD:
  Type: {self.config['cicd']['type']}

Deployment:
  Auto-merge: {self.config['deployment']['auto_merge']}
  Monitor timeout: {self.config['deployment']['monitor_timeout_seconds']}s
  Max retries: {self.config['deployment']['max_retries']}
""")

        # Save
        print("\nSaving configuration...")
        self.save_config()

        self.log("WIZARD COMPLETE!", "success")
        print(f"""
Next steps:
  1. Make sure your GitHub token is set: export {self.config['github']['token_env_var']}=ghp_xxx
  2. Run: python3 ops_phoenix.py --dry-run
  3. Or run full cycle: python3 ops_phoenix.py --env prod --full-cycle

Config saved to: {CONFIG_FILE}
        """)

def main():
    wizard = ConfigWizard()

    if len(sys.argv) > 1:
        if sys.argv[1] == "--config":
            wizard.show_current_config()
        elif sys.argv[1] == "--reset":
            wizard.reset_config()
        elif sys.argv[1] == "--help":
            print(__doc__)
        else:
            print(f"Unknown option: {sys.argv[1]}")
            print("Use: config_wizard.py [--config|--reset|--help]")
    else:
        wizard.run_wizard()

if __name__ == "__main__":
    main()
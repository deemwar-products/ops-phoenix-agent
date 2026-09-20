"""Configuration management for Ops Phoenix."""

import json
import os
import sys
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

from ops_phoenix.threshold import DecisionEngine

APP_NAME = "ops-phoenix"
CONFIG_DIR = Path.home() / f'.{APP_NAME}'

DEFAULTS: Dict[str, Any] = {
 "observability": {
  "type": "grafana_cloud",
  "grafana_url": "",
  "grafana_token_path": "~/Downloads/Archive/keys/grafana-api-token",
  "loki_url": "",
  "loki_user": "",
  "loki_password": "",
  "container_patterns": ["app-*", "api-*"],
  "host_patterns": ["*"],
  "error_patterns": ['level=~"(?i)error|fatal"'],
 },
 "github": {
  "repo": "",
  "token_env_var": "GITHUB_TOKEN",
  "base_branch": "main",
  "labels": ["ops-alert"],
 },
 "ai": {
  "provider": "anthropic",
  "model": "claude-sonnet-4-20250514",
  "api_key_env": "ANTHROPIC_API_KEY",
  "api_url": "https://api.anthropic.com",
 },
 "cicd": {
  "type": "github_actions",
  "workflow_name": "Deploy to Production",
 },
 "deployment": {
  "auto_merge": False,
  "monitor_timeout": 300,
  "max_retries": 2,
 },
 "agent": {
  "work_repo_path": str(CONFIG_DIR / "work" / "repo"),
  "test_dirs": {
   "go": ["apps/api"],
   "typescript": ["apps/ui", "apps/home", "apps/extension"],
  },
 },
 "thresholds": {},
}

class ConfigError(Exception):
 """Configuration validation error."""
 pass

class ConfigManager:
 """Load, validate, and manage Ops Phoenix configuration."""
 
 def __init__(self, config_path=None):
  self.config_path = config_path or self._find_config()
  self.config = {}
  self._load()
 
 def _find_config(self):
  """Search for config file in standard locations."""
  default_path = CONFIG_DIR / 'config.json'
  if default_path.exists():
   return default_path
  yaml_path = CONFIG_DIR / 'config.yaml'
  if yaml_path.exists():
   return yaml_path
  return None
 
 def _load(self):
  """Load config from file if it exists."""
  if self.config_path and self.config_path.exists():
   try:
    text = self.config_path.read_text()
    if self.config_path.suffix == '.yaml':
     try:
      import yaml
      self.config = yaml.safe_load(text) or {}
     except ImportError:
      raise ConfigError('PyYAML required for .yaml config files')
    else:
     self.config = json.loads(text)
   except (json.JSONDecodeError, ValueError) as e:
    raise ConfigError(f'Invalid config file: {e}')
  else:
   self.config = {}
 
 def needs_setup(self):
  """Check if required config values are missing."""
  required = ["github.repo", "ai.api_key_env"]
  return not all(self._nested_get(k) for k in required)
 
 def _nested_get(self, path):
  """Get a nested config value by dot path."""
  parts = path.split('.')
  val = self.config
  for p in parts:
   if isinstance(val, dict):
    val = val.get(p)
   else:
    return None
  return val
 
 def get(self, path, default=None):
  val = self._nested_get(path)
  return val if val is not None else default
 
 def set(self, path, value):
  """Set a nested config value by dot path."""
  parts = path.split('.')
  d = self.config
  for p in parts[:-1]:
   d = d.setdefault(p, {})
   d[parts[-1]] = value
 
 def validate(self):
  """Validate config and return list of error messages."""
  errors = []
  repo = self._nested_get('github.repo')
  if not repo or '/' not in repo:
   errors.append("github.repo is required (format: owner/repo)")
  api_key_env = self._nested_get('ai.api_key_env')
  if not api_key_env:
   errors.append("ai.api_key_env is required")
  elif not os.environ.get(api_key_env):
   errors.append(f'Environment variable {api_key_env} is not set')
  obs_type = self._nested_get('observability.type')
  if obs_type not in ("grafana_cloud", "loki_direct", "grafana_self_hosted", ""):
   errors.append("observability.type must be grafana_cloud, loki_direct, or grafana_self_hosted")
  if obs_type in ('grafana_cloud', 'grafana_self_hosted'):
   token_path = self._nested_get('observability.grafana_token_path')
   if not token_path:
    errors.append(f'grafana_token_path required for {obs_type}')
  if obs_type == 'loki_direct':
   loki_url = self._nested_get('observability.loki_url')
   if not loki_url:
    errors.append("observability.loki_url is required for loki_direct")
  return errors
 
 def with_defaults(self):
  """Return config merged with defaults."""
  result = {}
  for key, default_val in DEFAULTS.items():
   if key not in self.config:
    result[key] = dict(default_val) if isinstance(default_val, dict) else default_val
   else:
    val = self.config[key]
    if isinstance(default_val, dict) and isinstance(val, dict):
     merged = dict(default_val)
     merged.update(val)
     result[key] = merged
    else:
     result[key] = val
  for key, val in self.config.items():
   if key not in result:
    result[key] = val
  return result
 
 def save(self):
  """Save current config to the config file."""
  CONFIG_DIR.mkdir(parents=True, exist_ok=True)
  if self.config_path is None:
   self.config_path = CONFIG_DIR / 'config.json'
  self.config_path.write_text(json.dumps(self.config, indent=2))
 
 def show(self):
  """Return human-readable config summary."""
  return json.dumps(self._redact(self.config), indent=2)
 
 @staticmethod
 def _redact(obj):
  """Recursively redact sensitive values."""
  if isinstance(obj, dict):
   out = {}
   for k, v in obj.items():
    if any(s in k.lower() for s in ('token', 'key', 'password', 'secret')):
     out[k] = "(redacted)" if v else "(not set)"
    else:
     out[k] = ConfigManager._redact(v)
   return out
  if isinstance(obj, list):
   return [ConfigManager._redact(i) for i in obj]
  return obj
 
 def run_wizard(self):
  """Interactive setup wizard."""
  print("")
  print("=" * 60)
  print(" OPS PHOENIX - SETUP WIZARD")
  print("=" * 60)
  print("")
  print("I will ask you for the information needed.")
  print("Press Enter to accept defaults shown in [brackets].")
  print("")
  print("-" * 60)
  print(" STEP 1: Observability (logs)")
  print("-" * 60)
  print("")
  obs_type = self._ask(
   "What logging system do you use?",
   default=self.config.get("observability", {}).get("type", "grafana_cloud"),
   options=["grafana_cloud", "loki_direct", "grafana_self_hosted"],
  )
  self.set('observability.type', obs_type)
  
  if obs_type in ('grafana_cloud', 'grafana_self_hosted'):
   self.set('observability.grafana_url', self._ask(
    "Grafana URL",
    default=self.config.get("observability", {}).get("grafana_url", ""),
    required=True
   ))
   self.set('observability.grafana_token_path', self._ask(
    "Path to Grafana API token file",
    default=self.config.get("observability", {}).get("grafana_token_path", ""),
   ))
  if obs_type == 'grafana_cloud':
   self.set('observability.loki_url', '')
   self.set('observability.loki_user', '')
   self.set('observability.loki_password', '')
  
  elif obs_type == 'loki_direct':
   self.set('observability.loki_url', self._ask(
    "Loki URL (e.g., http://loki:3100)",
    default=self.config.get("observability", {}).get("loki_url", "http://localhost:3100"),
    required=True
   ))
   self.set('observability.loki_user', self._ask('Loki username (blank for none)'))
   self.set('observability.loki_password', self._ask('Loki password (blank for none)'))
  
  containers = self._ask(
   "Container patterns (comma-separated regex)",
   default=",".join(self.config.get("observability", {}).get("container_patterns", ["app-*", "api-*"])),
  )
  self.set('observability.container_patterns', [c.strip() for c in containers.split(',') if c.strip()])
  
  hosts = self._ask(
   "Host patterns (comma-separated regex)",
   default=",".join(self.config.get("observability", {}).get("host_patterns", ["*"])),
  )
  self.set('observability.host_patterns', [h.strip() for h in hosts.split(',') if h.strip()])
  
  print("")
  print("-" * 60)
  print(" STEP 2: GitHub (code + PRs)")
  print("-" * 60)
  print("")
  self.set('github.repo', self._ask(
   "GitHub repository (owner/repo)",
   default=self.config.get("github", {}).get("repo", ""),
   required=True
  ))
  self.set('github.base_branch', self._ask(
   "Default branch for PRs",
   default=self.config.get("github", {}).get("base_branch", "main"),
  ))
  self.set('github.token_env_var', self._ask(
   "GitHub token environment variable name",
   default=self.config.get("github", {}).get("token_env_var", "GITHUB_TOKEN"),
  ))
  
  labels = self._ask(
   "Labels for issues/PRs (comma-separated)",
   default=",".join(self.config.get("github", {}).get("labels", ["ops-alert"])),
  )
  self.set('github.labels', [l.strip() for l in labels.split(',') if l.strip()])
  
  print("")
  print("-" * 60)
  print(" STEP 3: AI (Claude)")
  print("-" * 60)
  print("")
  api_key_env = self._ask(
   "Claude API key environment variable",
   default=self.config.get("ai", {}).get("api_key_env", "ANTHROPIC_API_KEY"),
  )
  self.set('ai.api_key_env', api_key_env)
  
  api_url = self._ask(
   "Anthropic API base URL",
   default=self.config.get("ai", {}).get("api_url", "https://api.anthropic.com"),
  )
  self.set('ai.api_url', api_url)
  
  model = self._ask(
   "Claude model",
   default=self.config.get("ai", {}).get("model", "claude-sonnet-4-20250514"),
  )
  self.set('ai.model', model)
  
  print("")
  print("-" * 60)
  print(" STEP 4: CI/CD (deployments)")
  print("-" * 60)
  print("")
  cicd_type = self._ask(
   "Which CI/CD system triggers deployments?",
   default=self.config.get("cicd", {}).get("type", "github_actions"),
   options=["github_actions", "gitlab_ci", "jenkins", "manual"],
  )
  self.set('cicd.type', cicd_type)
  
  if cicd_type == 'github_actions':
   self.set('cicd.workflow_name', self._ask(
    "GitHub Actions workflow name for deployment",
    default=self.config.get("cicd", {}).get("workflow_name", "Deploy to Production"),
   ))
  
  print("")
  print("-" * 60)
  print(" STEP 5: Deployment settings")
  print("-" * 60)
  print("")
  auto_merge = self._ask(
   "Auto-merge PRs and deploy? (yes/no)",
   default="yes" if self.config.get("deployment", {}).get("auto_merge", False) else "no",
   options=["yes", "no"],
  )
  self.set('deployment.auto_merge', auto_merge == 'yes')
  
  timeout = self._ask(
   "Workflow monitoring timeout (seconds)",
   default=str(self.config.get("deployment", {}).get("monitor_timeout", 300)),
  )
  self.set('deployment.monitor_timeout', int(timeout))
  
  retries = self._ask(
   "Max retries on deploy failure",
   default=str(self.config.get("deployment", {}).get("max_retries", 2)),
  )
  self.set('deployment.max_retries', int(retries))
  
  print("")
  print("-" * 60)
  print(" STEP 6: Notification channels")
  print("-" * 60)
  print("")
  teams_enabled = self._ask(
   "Enable Microsoft Teams notifications? (yes/no)",
   default="yes" if self.config.get("notifications", {}).get("teams", {}).get("webhook_url") else "no",
   options=["yes", "no"],
  )
  self.set('notifications.teams.enabled', teams_enabled == 'yes')
  
  if teams_enabled == 'yes':
   self.set('notifications.teams.webhook_url', self._ask(
    "Power Automate flow URL",
    default=self.config.get("notifications", {}).get("teams", {}).get("webhook_url", ""),
    required=True
   ))
  
  telegram_enabled = self._ask(
   "Enable Telegram notifications? (yes/no)",
   default="yes" if self.config.get("notifications", {}).get("telegram", {}).get("bot_token") else "no",
   options=["yes", "no"],
  )
  self.set('notifications.telegram.enabled', telegram_enabled == 'yes')
  
  if telegram_enabled == 'yes':
   self.set('notifications.telegram.bot_token', self._ask(
    "Telegram Bot Token",
    default=self.config.get("notifications", {}).get("telegram", {}).get("bot_token", ""),
    required=True
   ))
   self.set('notifications.telegram.chat_id', self._ask(
    "Telegram Chat ID",
    default=self.config.get("notifications", {}).get("telegram", {}).get("chat_id", ""),
    required=True
   ))
  
  print("")
  print("=" * 60)
  print(" CONFIGURATION SUMMARY")
  print("=" * 60)
  print("")
  obs = self.config.get('observability', {})
  print(f"Observability: {obs.get('type', 'not set')}")
  print(f" Containers: {', '.join(obs.get('container_patterns', []))}")
  if obs.get('grafana_url'):
   print(f" Grafana: {obs['grafana_url']}")
  if obs.get('loki_url'):
   print(f" Loki: {obs['loki_url']}")
  
  gh = self.config.get('github', {})
  print(f"GitHub: {gh.get('repo', 'not set')}")
  print(f" Branch: {gh.get('base_branch', 'main')}")
  print(f" Labels: {', '.join(gh.get('labels', []))}")
  
  ai = self.config.get('ai', {})
  print(f"AI: model={ai.get('model', 'not set')}")
  print(f" API key env: {ai.get('api_key_env', 'not set')}")
  
  cicd = self.config.get('cicd', {})
  print(f"CI/CD: {cicd.get('type', 'not set')}")
  
  dep = self.config.get('deployment', {})
  print(f"Deployment: auto_merge={dep.get('auto_merge', False)}, timeout={dep.get('monitor_timeout', 300)}s")
  
  notif = self.config.get('notifications', {})
  teams_cfg = notif.get('teams', {})
  telegram_cfg = notif.get('telegram', {})
  print("Notifications:")
  print(f" Teams: {'enabled' if teams_cfg.get('enabled') else 'disabled'}{' (configured)' if teams_cfg.get('webhook_url') else ''}")
  print(f" Telegram: {'enabled' if telegram_cfg.get('enabled') else 'disabled'}{' (configured)' if telegram_cfg.get('bot_token') else ''}")
  
  print(f"Saving to {self.config_path} ...")
  self.save()
  print("Configuration saved!")
  return True
 
 def _ask(self, question, default='', required=False, options=None):
  """Prompt the user for input."""
  if options:
   print(f"")
   print(f"{question}")
   for i, opt in enumerate(options, 1):
    marker = ' (default)' if opt == default else ''
    print(f" {i}. {opt}{marker}")
   while True:
    choice = input(f'Enter choice [1-{len(options)}]: ').strip()
    if not choice and default:
     return default
    try:
     idx = int(choice) - 1
     if 0 <= idx < len(options):
      return options[idx]
    except ValueError:
     pass
    print("Please enter 1-" + str(len(options)))
  else:
   prompt = f"\n{question}"
   if default:
    prompt += f" [{default}]"
   if required:
    prompt += ' *'
   prompt += ': '
   while True:
    answer = input(prompt).strip()
    if answer:
     return answer
    if default:
     return default
    if not required:
     return ''
    print("This field is required")
    return ''

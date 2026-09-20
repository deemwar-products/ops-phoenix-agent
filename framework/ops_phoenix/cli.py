"""CLI entry point for Ops Phoenix."""

import argparse
import json
import sys
from pathlib import Path

from ops_phoenix.agent import OpsPhoenix
from ops_phoenix.config import ConfigManager, CONFIG_DIR
from ops_phoenix.version import __version__

def main():
 parser = argparse.ArgumentParser(
  prog="ops-phoenix",
  description="Autonomous ops agent: detect, analyze, fix, and deploy.",
 )
 parser.add_argument(
  "--version", action="version", version=f"ops-phoenix {__version__}"
 )
 sub = parser.add_subparsers(dest='command', help='Available commands')
 
 # init
 p_init = sub.add_parser('init', help='Run the interactive setup wizard')
 
 # run
 p_run = sub.add_parser('run', help='Run the full detect-analyze-fix-deploy cycle')
 p_run.add_argument('--time-window', default='1h', help='Time window for log query (e.g. 30m, 2h)')
 p_run.add_argument('--mode', choices=['dev', 'prod'], default='dev', help='Run mode')
 
 # status
 p_status = sub.add_parser('status', help='Show config and connection status')
 
 # test
 p_test = sub.add_parser('test', help='Test all configured connections')
 
 # show
 p_show = sub.add_parser('show', help='Show current configuration (secrets redacted)')
 
 args = parser.parse_args()
 if not args.command:
  parser.print_help()
  sys.exit(1)
 
 try:
  _dispatch(args)
 except KeyboardInterrupt:
  print(chr(10) + "Interrupted.", file=sys.stderr)
  sys.exit(130)
 except Exception as e:
  print(f"ERROR: {e}", file=sys.stderr)
  sys.exit(1)

def _dispatch(args):
 """Route command to handler."""
 if args.command == 'init':
  _cmd_init()
 elif args.command == 'run':
  _cmd_run(args)
 elif args.command == 'status':
  _cmd_status()
 elif args.command == 'test':
  _cmd_test()
 elif args.command == 'show':
  _cmd_show()

def _cmd_init():
 """Run interactive setup wizard."""
 mgr = ConfigManager()
 ok = mgr.run_wizard()
 sys.exit(0 if ok else 1)

def _cmd_run(args):
 """Run the full ops cycle."""
 agent = OpsPhoenix()
 agent.mode = args.mode
 result = agent.run_full_cycle(time_window=args.time_window)
 print(json.dumps(result, indent=2))
 sys.exit(0 if result.get('status') == 'success' else 1)

def _cmd_status():
 """Show config and connection status."""
 mgr = ConfigManager()
 print("Configuration:")
 print(mgr.show())
 print("")
 if mgr.needs_setup():
  print("NOT SET UP - run: ops-phoenix init", file=sys.stderr)
  sys.exit(1)
 agent = OpsPhoenix(config=mgr)
 ok = agent.test_connections()
 sys.exit(0 if ok else 1)

def _cmd_test():
 """Test all configured connections."""
 agent = OpsPhoenix()
 ok = agent.test_connections()
 print(chr(10) + "All connections OK" if ok else chr(10) + "Some connections FAILED", file=sys.stderr)
 sys.exit(0 if ok else 1)

def _cmd_show():
 """Show current configuration."""
 mgr = ConfigManager()
 errors = mgr.validate()
 if errors:
  print("Configuration errors:", file=sys.stderr)
  for e in errors:
   print(f" - {e}", file=sys.stderr)
  sys.exit(1)
 print(mgr.show())

if __name__ == '__main__':
 main()

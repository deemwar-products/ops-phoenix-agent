"""Smoke tests for the framework/ops_phoenix.py CLI surface changes (spec 055 slice 3).

These cover the new --action auto-cycle handler and the --config <path> flag
that the autonomous runner wrapper on vm-installer depends on. The end-to-end
behaviour (real GitHub/Grafana calls) lives in the workflow scenarios; here we
only check the dispatch wires up correctly.

These tests are written as stdlib-only so they run on any Python with the
framework in sys.path; the existing project test runner may also pick them up
via pytest.
"""

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

# Make the framework importable when unittest runs from anywhere.
FRAMEWORK_DIR = Path(__file__).parent.parent
sys.path.insert(0, str(FRAMEWORK_DIR))

from ops_phoenix import OpsPhoenix # noqa: E402


def _make_config(tmp_path):
  cfg = {"github": {"repo": "deemwar-products/reqsume"}, "grafana": {"url": "http://x"}}
  p = Path(tmp_path) / "config.json"
  p.write_text(json.dumps(cfg))
  return p


class LoadConfigTests(unittest.TestCase):
  def test_load_config_accepts_explicit_path(self):
    with tempfile.TemporaryDirectory() as d:
      cfg_path = _make_config(d)
      agent = OpsPhoenix(config_path=cfg_path)
      self.assertEqual(
        agent.config.get("github", {}).get("repo"),
        "deemwar-products/reqsume",
      )

  def test_load_config_defaults_when_no_path(self):
    # When neither config_path nor an env var is provided, and the
    # CONFIG_FILE module-level constant points at a missing file,
    # load_config should leave self.config as the empty dict that
    # __init__ created. We approximate this by passing an explicit
    # nonexistent path: load_config("...") returns the existing dict.
    agent = OpsPhoenix(config_path=Path("/nonexistent/does-not-exist.json"))
    self.assertEqual(agent.config, {})

  def test_load_config_picks_up_env_var(self):
    with tempfile.TemporaryDirectory() as d:
      cfg_path = _make_config(d)
      with mock.patch.dict(
        os.environ,
        {"OPS_PHOENIX_CONFIG": str(cfg_path), "HOME": d},
      ):
        agent = OpsPhoenix()
        self.assertEqual(
          agent.config.get("github", {}).get("repo"),
          "deemwar-products/reqsume",
        )


class RunAutoCycleTests(unittest.TestCase):
  def test_run_auto_cycle_method_exists(self):
    agent = OpsPhoenix()
    self.assertTrue(callable(getattr(agent, "run_auto_cycle", None)))

  def test_run_auto_cycle_returns_dict_on_healthy(self):
    with tempfile.TemporaryDirectory() as d:
      with mock.patch.dict(os.environ, {"HOME": d}):
        agent = OpsPhoenix()

    def fake_test_connections():
      return True

    def fake_detect_errors(_):
      return []

    with mock.patch.object(agent, "test_connections", fake_test_connections), \
      mock.patch.object(agent, "detect_errors", fake_detect_errors):
        result = agent.run_auto_cycle()
        self.assertEqual(result["status"], "healthy")

"""Grafana Cloud adapter - queries Loki via Grafana's datasource API."""

import json
import os
import subprocess
from pathlib import Path
from typing import Any, Dict, List, Optional

from ops_phoenix.adapters.base import ObservabilityAdapter


class GrafanaCloudAdapter(ObservabilityAdapter):
    """Query Grafana Cloud Loki via the API."""

    def __init__(self, config: dict):
        super().__init__(config)
        self.url = config.get("url", "https://logs-prod-us-central1.grafana.net")
        self.api_key = config.get("api_key", "") or os.environ.get("GRAFANA_API_KEY", "")
        self.org_id = config.get("org_id", "1")
        self.datasource = config.get("datasource", "logs")

    def query_errors(self, time_window: str = "1h") -> list[str]:
        """Query Loki for error logs."""
        from datetime import datetime, timedelta
        import time as time_mod
        import urllib.request

        end_ns = time_mod.time_ns()
        start = datetime.now() - timedelta(hours=1)
        start_ns = int(start.timestamp() * 1e9)
        duration_ns = end_ns - start_ns

        expr = '{{severity=~"(?i)error|fatal|panic"}}'
        expr += self.get_container_filter()

        url = (
            f"{self.url}/loki/api/v1/query_range"
            f"?query={urllib.parse.quote(expr)}"
            f"&start={start_ns}&end={end_ns}&step=60"
        )

        req = urllib.request.Request(url)
        req.add_header("Authorization", f"Bearer {self.api_key}")
        req.add_header("X-Scope-OrgID", self.org_id)

        try:
            import urllib.parse
            with urllib.request.urlopen(req, timeout=30) as resp:
                data = json.loads(resp.read())
                results = data.get("data", {}).get("result", [])
                return [str(r) for r in results]
        except Exception as e:
            print(f"Grafana Cloud query failed: {e}")
            return []

    def test_connection(self) -> bool:
        """Check if Grafana Cloud is reachable."""
        try:
            import urllib.request
            url = f"{self.url}/api/health"
            req = urllib.request.Request(url)
            req.add_header("Authorization", f"Bearer {self.api_key}")
            with urllib.request.urlopen(req, timeout=10) as resp:
                return resp.status == 200
        except Exception:
            return False

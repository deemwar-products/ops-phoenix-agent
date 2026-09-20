"""Loki direct query adapter."""

import json
import os
from typing import Any, Dict, List

from ops_phoenix.adapters.base import ObservabilityAdapter


class LokiDirectAdapter(ObservabilityAdapter):
    """Query Loki directly via HTTP API."""

    def __init__(self, config: dict):
        super().__init__(config)
        self.url = config.get("url", "http://localhost:3100")
        self.org_id = config.get("org_id", "1")
        self.tenant = config.get("tenant")

    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query Loki for error log lines."""
        import time as time_mod
        from datetime import datetime, timedelta
        import urllib.request, urllib.parse

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
        if self.tenant:
            req.add_header("X-Scope-OrgID", self.tenant)

        try:
            with urllib.request.urlopen(req, timeout=30) as resp:
                data = json.loads(resp.read())
                results = data.get("data", {}).get("result", [])
                return [str(r) for r in results]
        except Exception as e:
            print(f"Loki query failed: {e}")
            return []

    def test_connection(self) -> bool:
        """Check if Loki is reachable."""
        try:
            import urllib.request
            req = urllib.request.Request(f"{self.url}/ready")
            if self.tenant:
                req.add_header("X-Scope-OrgID", self.tenant)
            with urllib.request.urlopen(req, timeout=10) as resp:
                return resp.status == 200
        except Exception:
            return False

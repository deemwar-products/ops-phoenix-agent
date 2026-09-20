"""Grafana self-hosted / Loki adapter."""

import os
from typing import Any, Dict, List

from ops_phoenix.adapters.loki_direct import LokiDirectAdapter


class GrafanaSelfHostedAdapter(LokiDirectAdapter):
    """Self-hosted Grafana Loki instance."""

    def __init__(self, config: dict):
        super().__init__(config)
        self.grafana_url = config.get("grafana_url", self.url)
        self.api_key = config.get("api_key", "") or os.environ.get("GRAFANA_API_KEY", "")

    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query via Grafana's Explore API."""
        return super().query_errors(time_window)

    def test_connection(self) -> bool:
        """Check Grafana availability."""
        return super().test_connection()

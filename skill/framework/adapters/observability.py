#!/usr/bin/env python3
"""
Observability Adapters for Ops Phoenix

Provides a unified interface for querying logs from different providers.
"""

import json
import subprocess
from abc import ABC, abstractmethod
from typing import List, Dict, Optional
from pathlib import Path

class ObservabilityAdapter(ABC):
    """Base class for observability adapters"""

    def __init__(self, config: Dict):
        self.config = config

    @abstractmethod
    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query logs for errors"""
        pass

    @abstractmethod
    def test_connection(self) -> bool:
        """Test if the connection works"""
        pass

    def get_container_filter(self) -> str:
        """Get Loki container filter"""
        containers = self.config.get("container_patterns", ["app-*"])
        container_expr = "|".join(containers)
        return f'{{container=~"{container_expr}"}}'

    def get_host_filter(self) -> str:
        """Get Loki host filter"""
        hosts = self.config.get("host_patterns", ["prod-*"])
        host_expr = "|".join(hosts)
        return f'host=~"{host_expr}"'

    def get_error_pattern(self) -> str:
        """Get error pattern"""
        patterns = self.config.get("error_patterns", ['level=~"(?i)error|fatal"'])
        return patterns[0] if patterns else 'level=~"(?i)error|fatal"'


class GrafanaCloudAdapter(ObservabilityAdapter):
    """Adapter for Grafana Cloud with Loki"""

    def __init__(self, config: Dict):
        super().__init__(config)
        # Extract observability config
        obs_config = config.get("observability", config)  # Fallback to full config for backwards compat
        self.url = obs_config.get("grafana_url", "")
        self.token_path = Path(obs_config.get("grafana_token_path", "").replace("~", str(Path.home())))
        self.api_url = f"{self.url}/api/ds/query?ds_type=loki"

    def load_token(self) -> str:
        """Load Grafana API token"""
        # Expand ~ in path
        token_path = Path(self.token_path).expanduser()
        if not token_path.exists():
            raise FileNotFoundError(f"Grafana token not found: {token_path}")
        with open(token_path) as f:
            return f.read().strip()

    def get_time_range(self, window: str) -> tuple:
        """Map time window to Grafana duration strings"""
        ranges = {
            "1h": ("now-1h", "now"),
            "6h": ("now-6h", "now"),
            "24h": ("now-24h", "now"),
            "7d": ("now-7d", "now"),
        }
        return ranges.get(window, ("now-1h", "now"))

    def test_connection(self) -> bool:
        """Test Grafana Cloud connection"""
        try:
            token = self.load_token()
            # Simple health check
            result = subprocess.run([
                "curl", "-sS",
                "-H", f"Authorization: Bearer {token}",
                self.url + "/api/health"
            ], capture_output=True, text=True)
            return result.returncode == 0
        except Exception as e:
            print(f"Connection test failed: {e}")
            return False

    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query Loki for errors via Grafana Cloud"""
        token = self.load_token()
        from_ts, to_ts = self.get_time_range(time_window)

        # Build query
        container_filter = self.get_container_filter()
        host_filter = self.get_host_filter()
        error_pattern = self.get_error_pattern()

        query_expr = f'{container_filter}, {host_filter} | json | {error_pattern}'

        query = {
            "queries": [{
                "expr": query_expr,
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
            self.api_url,
            "--data", json.dumps(query)
        ], capture_output=True, text=True)

        if result.returncode != 0:
            print(f"Loki query failed: {result.stderr}")
            return []

        return self.parse_response(result.stdout)

    def parse_response(self, response: str) -> List[str]:
        """Parse Loki response"""
        try:
            data = json.loads(response)
            frames = data.get("results", {}).get("A", {}).get("frames", [])
            if not frames:
                return []

            vals = frames[0].get("data", {}).get("values", [])
            if len(vals) <= 2:
                return []

            return [l for l in vals[2] if l and l.strip()]
        except json.JSONDecodeError:
            return []


class LokiDirectAdapter(ObservabilityAdapter):
    """Adapter for direct Loki access (self-hosted)"""

    def __init__(self, config: Dict):
        super().__init__(config)
        obs_config = config.get("observability", config)
        self.url = obs_config.get("loki_url", "http://localhost:3100")
        self.user = obs_config.get("loki_user", "")
        self.password = obs_config.get("loki_password", "")

    def test_connection(self) -> bool:
        """Test Loki connection"""
        try:
            result = subprocess.run([
                "curl", "-sS",
                f"{self.url}/ready"
            ], capture_output=True, text=True)
            return "ready" in result.stdout.lower()
        except Exception:
            return False

    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query Loki directly"""
        container_filter = self.get_container_filter()
        host_filter = self.get_host_filter()
        error_pattern = self.get_error_pattern()

        # Build LogQL query
        query = f'{container_filter} | json | {error_pattern}'

        # Build curl command
        cmd = [
            "curl", "-sS",
            "-G", f"{self.url}/loki/api/v1/query_range",
            "--data-urlencode", f"query={query}",
            "-H", "Content-Type: application/json"
        ]

        if self.user and self.password:
            cmd.extend(["-u", f"{self.user}:{self.password}"])

        # Add time range
        time_map = {
            "1h": "3600",
            "6h": "21600",
            "24h": "86400",
            "7d": "604800"
        }
        seconds = time_map.get(time_window, "3600")

        result = subprocess.run(cmd + [
            "--data", f"limit=100&start=now-{seconds}s&end=now"
        ], capture_output=True, text=True)

        if result.returncode != 0:
            print(f"Loki query failed: {result.stderr}")
            return []

        return self.parse_response(result.stdout)

    def parse_response(self, response: str) -> List[str]:
        """Parse Loki LogQL response"""
        try:
            data = json.loads(response)
            results = data.get("data", {}).get("result", [])

            errors = []
            for result in results:
                values = result.get("values", [])
                for value in values:
                    if len(value) >= 2:
                        errors.append(value[1])

            return errors
        except json.JSONDecodeError:
            return []


class GrafanaSelfHostedAdapter(ObservabilityAdapter):
    """Adapter for self-hosted Grafana (proxies to Loki)"""

    def __init__(self, config: Dict):
        super().__init__(config)
        obs_config = config.get("observability", config)
        self.url = obs_config.get("grafana_url", "")
        self.token_path = Path(obs_config.get("grafana_token_path", "").replace("~", str(Path.home())))
        self.api_url = f"{self.url}/api/ds/query?ds_type=loki"

    def load_token(self) -> str:
        """Load Grafana API token"""
        # Expand ~ in path
        token_path = Path(self.token_path).expanduser()
        if not token_path.exists():
            raise FileNotFoundError(f"Grafana token not found: {token_path}")
        with open(token_path) as f:
            return f.read().strip()

    def test_connection(self) -> bool:
        """Test Grafana connection"""
        try:
            token = self.load_token()
            result = subprocess.run([
                "curl", "-sS",
                "-H", f"Authorization: Bearer {token}",
                f"{self.url}/api/health"
            ], capture_output=True, text=True)
            return result.returncode == 0
        except Exception:
            return False

    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query Loki via self-hosted Grafana"""
        token = self.load_token()

        time_map = {
            "1h": ("now-1h", "now"),
            "6h": ("now-6h", "now"),
            "24h": ("now-24h", "now"),
            "7d": ("now-7d", "now"),
        }
        from_ts, to_ts = time_map.get(time_window, ("now-1h", "now"))

        container_filter = self.get_container_filter()
        host_filter = self.get_host_filter()
        error_pattern = self.get_error_pattern()

        query_expr = f'{container_filter}, {host_filter} | json | {error_pattern}'

        query = {
            "queries": [{
                "expr": query_expr,
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
            self.api_url,
            "--data", json.dumps(query)
        ], capture_output=True, text=True)

        if result.returncode != 0:
            return []

        return self.parse_response(result.stdout)

    def parse_response(self, response: str) -> List[str]:
        """Parse Loki response via Grafana"""
        try:
            data = json.loads(response)
            frames = data.get("results", {}).get("A", {}).get("frames", [])
            if not frames:
                return []

            vals = frames[0].get("data", {}).get("values", [])
            if len(vals) <= 2:
                return []

            return [l for l in vals[2] if l and l.strip()]
        except json.JSONDecodeError:
            return []


def create_adapter(config: Dict) -> ObservabilityAdapter:
    """Factory to create the appropriate adapter"""
    obs_type = config.get("observability", {}).get("type", "grafana_cloud")

    adapters = {
        "grafana_cloud": GrafanaCloudAdapter,
        "loki_direct": LokiDirectAdapter,
        "grafana_self_hosted": GrafanaSelfHostedAdapter,
    }

    adapter_class = adapters.get(obs_type, GrafanaCloudAdapter)
    return adapter_class(config)
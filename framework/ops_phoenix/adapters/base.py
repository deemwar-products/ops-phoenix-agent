"""Base observability adapter."""

from abc import ABC, abstractmethod
from typing import List, Dict, Any


class ObservabilityAdapter(ABC):
    """Base class for observability log adapters."""

    def __init__(self, config: dict):
        self.config = config

    @abstractmethod
    def query_errors(self, time_window: str = "1h") -> List[str]:
        """Query logs for errors within the given time window."""
        ...

    @abstractmethod
    def test_connection(self) -> bool:
        """Verify the adapter can reach its data source."""
        ...

    def get_container_filter(self) -> str:
        """Build Loki container filter from config patterns."""
        containers = self.config.get("container_patterns", ["app-*"])
        expr = "|".join(containers)
        return f'{{{{container=~"{expr}"}}}}'

    def get_host_filter(self) -> str:
        """Build Loki host filter from config patterns."""
        hosts = self.config.get("host_patterns", ["*"])
        expr = "|".join(hosts)
        return f'host=~"{expr}"'

    def get_error_pattern(self) -> str:
        """Get the first configured error pattern."""
        patterns = self.config.get("error_patterns", ['level=~"(?i)error|fatal"'])
        return patterns[0] if patterns else 'level=~"(?i)error|fatal"'

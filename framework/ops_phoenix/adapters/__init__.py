"""Observability adapters for Ops Phoenix.

Provides a unified interface for querying error logs from different providers.
"""

from ops_phoenix.adapters.base import ObservabilityAdapter
from ops_phoenix.adapters.factory import create_adapter

__all__ = ["ObservabilityAdapter", "create_adapter"]

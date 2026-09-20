"""Adapter factory for creating observability adapters."""

from ops_phoenix.adapters.base import ObservabilityAdapter
from ops_phoenix.adapters.grafana_cloud import GrafanaCloudAdapter
from ops_phoenix.adapters.grafana_self_hosted import GrafanaSelfHostedAdapter
from ops_phoenix.adapters.loki_direct import LokiDirectAdapter


def create_adapter(config: dict) -> ObservabilityAdapter:
    """Create the appropriate adapter based on config."""
    obs_config = config.get("observability", {})
    provider = obs_config.get("type", "").lower()

    if provider == "grafana_cloud":
        return GrafanaCloudAdapter(obs_config)
    elif provider == "grafana_self_hosted":
        return GrafanaSelfHostedAdapter(obs_config)
    elif provider in ("loki", "loki_direct"):
        return LokiDirectAdapter(obs_config)
    else:
        raise ValueError(f"Unknown observability provider: {provider}")

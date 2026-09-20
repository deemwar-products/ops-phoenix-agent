"""Notification channels for Ops Phoenix.

Supported channels:
 - teams: Microsoft Teams via Incoming Webhook
 - telegram: Telegram Bot API
"""

from ops_phoenix.notify.teams import (
 send_teams_message,
 errors_detected_card,
 issue_created_card,
 pr_created_card,
 deploy_result_card,
)
from ops_phoenix.notify.telegram import (
 send_telegram_message,
 send_telegram_html,
)

__all__ = [
 "send_teams_message",
 "errors_detected_card",
 "issue_created_card",
 "pr_created_card",
 "deploy_result_card",
 "send_telegram_message",
 "send_telegram_html",
]

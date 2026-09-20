"""Telegram notification via Bot API."""

import json
import os
import subprocess
from typing import Any, Dict, List, Optional


def send_telegram_message(bot_token: str, chat_id: str, text: str, parse_mode: str = "Markdown") -> bool:
 """Send a text message to a Telegram chat."""
 
 url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
 payload = {
  "chat_id": chat_id,
  "text": text,
  "parse_mode": parse_mode,
 }
 
 result = subprocess.run(
  [
   "curl", "-sS", "-X", "POST",
   "-H", "Content-Type: application/json",
   "--data", json.dumps(payload),
   url,
  ],
  capture_output=True,
  text=True,
  timeout=15,
  )
 
 output = json.loads(result.stdout) if result.stdout else {}
 if output.get("ok"):
  return True
 
 print(f"Telegram send failed: {output}")
 return False


def send_telegram_html(bot_token: str, chat_id: str, html: str) -> bool:
 """Send an HTML-formatted message to a Telegram chat."""
 
 url = f"https://api.telegram.org/bot{bot_token}/sendMessage"
 payload = {
  "chat_id": chat_id,
  "text": html,
  "parse_mode": "HTML",
 }
 
 result = subprocess.run(
  [
   "curl", "-sS", "-X", "POST",
   "-H", "Content-Type: application/json",
   "--data", json.dumps(payload),
   url,
  ],
  capture_output=True,
  text=True,
  timeout=15,
  )
 
 output = json.loads(result.stdout) if result.stdout else {}
 if output.get("ok"):
  return True
 
 print(f"Telegram send failed: {output}")
 return False


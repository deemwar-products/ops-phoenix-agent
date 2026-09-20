"""AI-powered error analysis using Claude API."""

import os
import time
from typing import Any, Dict, List, Optional


def analyze_errors(
    errors: List[str],
    api_key: str,
    api_url: str = "https://api.anthropic.com",
    model: str = "claude-sonnet-4-20250514",
    max_retries: int = 3,
) -> Optional[Dict[str, Any]]:
    """Send errors to Claude for root-cause analysis."""
    if not errors or not api_key:
        return None

    error_sample = "\n".join(errors[:20])  # Limit to 20 errors for context
    prompt = f"""You are an SRE expert. Analyze these error logs and provide:

1. Root cause (one sentence)
2. Severity: critical / high / medium / low
3. Affected component(s)
4. Suggested fix
5. Files likely affected (if identifiable)

Errors:
```
{error_sample}
```

Respond in JSON format:
{{
  "root_cause": "...",
  "severity": "...",
  "affected": "...",
  "fix_suggestion": "...",
  "files_affected": ["..."]
}}
"""

    for attempt in range(max_retries):
        try:
            import urllib.request, json as json_mod
            payload = json_mod.dumps({
                "model": model,
                "max_tokens": 1024,
                "messages": [{"role": "user", "content": prompt}],
            }).encode()

            req = urllib.request.Request(
                f"{api_url}/v1/messages",
                data=payload,
                headers={
                    "x-api-key": api_key,
                    "anthropic-version": "2023-06-01",
                    "content-type": "application/json",
                },
                method="POST",
            )

            with urllib.request.urlopen(req, timeout=60) as resp:
                result = json_mod.loads(resp.read())
                text = result["content"][0]["text"]

                # Extract JSON from response
                import re
                m = re.search(r'\{.*\}', text, re.DOTALL)
                if m:
                    return json_mod.loads(m.group())

            return {"raw_response": text}
        except Exception as e:
            if attempt < max_retries - 1:
                time.sleep(2 ** attempt)
            else:
                print(f"AI analysis failed after {max_retries} attempts: {e}")
                return None
    return None


def generate_fix_diff(
    error_summary: str,
    analysis: Dict[str, Any],
    source_context: str,
    api_key: str,
    api_url: str = "https://api.anthropic.com",
    model: str = "claude-sonnet-4-20250514",
) -> Optional[str]:
    """Generate a code fix as a unified diff."""
    prompt = f"""Given this error analysis, generate a unified diff (patch) that fixes the issue.

Error summary:
{error_summary}

Root cause: {analysis.get('root_cause', 'N/A')}
Fix suggestion: {analysis.get('fix_suggestion', 'N/A')}

Relevant source code:
{source_context[:5000]}

Respond with ONLY a unified diff patch. No explanations.
"""

    try:
        import urllib.request, json as json_mod, re
        payload = json_mod.dumps({
            "model": model,
            "max_tokens": 2048,
            "messages": [{"role": "user", "content": prompt}],
        }).encode()

        req = urllib.request.Request(
            f"{api_url}/v1/messages",
            data=payload,
            headers={
                "x-api-key": api_key,
                "anthropic-version": "2023-06-01",
                "content-type": "application/json",
            },
            method="POST",
        )

        with urllib.request.urlopen(req, timeout=120) as resp:
            result = json_mod.loads(resp.read())
            text = result["content"][0]["text"]

            # Extract diff
            m = re.search(r'(--- .+\n\+\+\+ .+\n.*)', text, re.DOTALL)
            return m.group(1) if m else text

    except Exception as e:
        print(f"Diff generation failed: {e}")
        return None

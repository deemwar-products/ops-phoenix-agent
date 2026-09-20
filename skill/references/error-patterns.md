# Error Patterns Reference

Common error patterns to detect in production logs.

## Critical (Immediate Alert)

| Pattern | Description | Action |
|---------|-------------|--------|
| `FATAL` | Application crash | Create issue + page |
| `PANIC` | Go panic | Create issue + page |
| `Out of memory` | OOM kill | Create issue + page |

## Warning (Create Issue)

| Pattern | Description | Action |
|---------|-------------|--------|
| `ERROR` | General error | Create issue |
| `500` | HTTP 500 | Create issue |
| `503` | HTTP 503 (service unavailable) | Create issue |
| `validation_failed` | User input validation failed | Create issue |
| `extraction_failed` | AI extraction failed | Create issue |
| `timeout\|TIMEOUT` | Operation timed out | Create issue |
| `connection refused` | Network connection failed | Create issue |

## Info (Log Only)

| Pattern | Description | Action |
|---------|-------------|--------|
| `404` | Not found (normal) | Log only |
| `context canceled` | Request cancelled (normal) | Log only |
| `EOF` | End of stream (normal) | Log only |

## Known Error Types

### Profile Creation

```
validation_failed - user profile creation
```

### AI/LLM Errors

```
Gemini API timeout
OpenAI rate limit
extraction_failed
```

### Database Errors

```
connection refused
pq: timeout
context deadline exceeded
```

### Authentication

```
firebase token invalid
jwt expired
```

## Aggregation Rules

Group errors by:
1. First 50 chars of error message
2. Endpoint/request type
3. User ID (if present)

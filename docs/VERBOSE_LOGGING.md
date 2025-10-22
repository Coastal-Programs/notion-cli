# Verbose Logging Guide

This document describes the structured event logging system for observability and debugging in notion-cli.

## Overview

Verbose logging provides machine-readable, structured JSON events to **stderr only**, ensuring stdout JSON output remains clean for automation. This is critical for AI agents and scripts that parse stdout.

## Key Features

- **Structured JSON format** - Easy to parse with `jq`, `grep`, or log aggregators
- **Stderr only** - Never pollutes stdout JSON output
- **Zero performance impact when disabled** - No overhead in production
- **Configurable verbosity** - Enable via flag or environment variable
- **Event types** - Retry events, cache events, rate limiting, circuit breaker

## Enabling Verbose Logging

### Method 1: Command-line Flag

```bash
# Enable verbose logging for a single command
notion-cli page retrieve PAGE_ID --json --verbose 2>debug.log

# Redirect stderr to file, keep stdout for JSON
notion-cli db query DS_ID --json --verbose 2>debug.log >output.json

# View only verbose logs (stderr)
notion-cli search -q "test" --json --verbose 2>&1 >/dev/null
```

### Method 2: Environment Variable

```bash
# Enable verbose logging for all commands in session
export NOTION_CLI_VERBOSE=true
notion-cli page retrieve PAGE_ID --json

# Windows CMD
set NOTION_CLI_VERBOSE=true

# Windows PowerShell
$env:NOTION_CLI_VERBOSE="true"
```

### Method 3: DEBUG Environment Variable

```bash
# Enable all debug features (includes verbose logging)
export DEBUG=true
notion-cli page retrieve PAGE_ID --json

# Alternative
export NOTION_CLI_DEBUG=true
```

## Event Types

### 1. Retry Events

#### Retry Attempt
Logged when a retry is being attempted after a failure.

```json
{
  "level": "warn",
  "event": "retry",
  "attempt": 2,
  "max_retries": 3,
  "reason": "SERVICE_UNAVAILABLE",
  "retry_after_ms": 2000,
  "url": "/v1/databases",
  "context": "fetchDataSource",
  "status_code": 503,
  "error_code": "service_unavailable",
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Fields:**
- `level`: Severity (`warn` for retriable errors)
- `event`: Always `"retry"` for retry attempts
- `attempt`: Current retry attempt number (1-indexed)
- `max_retries`: Maximum number of retries configured
- `reason`: High-level error category (see Error Reasons below)
- `retry_after_ms`: Delay before next retry in milliseconds
- `url`: API endpoint or operation context
- `context`: Function/operation name
- `status_code`: HTTP status code (if applicable)
- `error_code`: Notion API error code (if applicable)
- `timestamp`: ISO 8601 timestamp

#### Rate Limited
Special event for rate limiting (HTTP 429).

```json
{
  "level": "warn",
  "event": "rate_limited",
  "attempt": 1,
  "max_retries": 3,
  "reason": "RATE_LIMITED",
  "retry_after_ms": 1200,
  "url": "/v1/databases/abc123/query",
  "context": "queryDatabase",
  "status_code": 429,
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Use Case:** Monitor rate limit frequency to optimize API usage patterns.

#### Retry Exhausted
Logged when all retries are exhausted and operation fails.

```json
{
  "level": "error",
  "event": "retry_exhausted",
  "attempt": 4,
  "max_retries": 3,
  "reason": "SERVICE_UNAVAILABLE",
  "context": "fetchDataSource",
  "status_code": 503,
  "error_code": "service_unavailable",
  "timestamp": "2025-10-23T14:32:20.456Z"
}
```

**Use Case:** Alert on persistent failures requiring manual intervention.

#### Retry Attempt Start
Logged when starting a retry attempt (not first attempt).

```json
{
  "level": "info",
  "event": "retry_attempt",
  "attempt": 2,
  "max_retries": 3,
  "context": "fetchDataSource",
  "timestamp": "2025-10-23T14:32:17.234Z"
}
```

### 2. Cache Events

#### Cache Hit
Logged when data is successfully retrieved from cache.

```json
{
  "level": "debug",
  "event": "cache_hit",
  "namespace": "dataSource",
  "key": "abc123",
  "age_ms": 5234,
  "ttl_ms": 600000,
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Fields:**
- `namespace`: Cache namespace (dataSource, page, block, user)
- `key`: Cache key (usually resource ID)
- `age_ms`: Age of cached data in milliseconds
- `ttl_ms`: Time-to-live for this cache entry

**Use Case:** Calculate cache hit rate and effectiveness.

#### Cache Miss
Logged when data is not found in cache or expired.

```json
{
  "level": "debug",
  "event": "cache_miss",
  "namespace": "page",
  "key": "xyz789",
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Use Case:** Identify frequently requested uncached data.

#### Cache Set
Logged when data is stored in cache.

```json
{
  "level": "debug",
  "event": "cache_set",
  "namespace": "dataSource",
  "key": "abc123",
  "ttl_ms": 600000,
  "cache_size": 45,
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Fields:**
- `cache_size`: Current number of entries in cache

**Use Case:** Monitor cache growth and memory usage.

#### Cache Invalidation
Logged when cache entries are manually invalidated.

```json
{
  "level": "debug",
  "event": "cache_invalidate",
  "namespace": "dataSource",
  "key": "abc123",
  "cache_size": 44,
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Use Case:** Track cache invalidation patterns after writes.

#### Cache Eviction
Logged when cache entries are evicted (LRU or expired).

```json
{
  "level": "debug",
  "event": "cache_evict",
  "namespace": "lru",
  "key": "dataSource:old123",
  "cache_size": 999,
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Use Case:** Detect cache thrashing or inadequate cache size.

### 3. Circuit Breaker Events

#### Circuit Breaker Opened
Logged when circuit breaker opens due to repeated failures.

```json
{
  "level": "error",
  "event": "retry_exhausted",
  "attempt": 5,
  "max_retries": 5,
  "reason": "CIRCUIT_OPENED",
  "retry_after_ms": 60000,
  "context": "Circuit breaker opened after 5 failures",
  "timestamp": "2025-10-23T14:32:15.234Z"
}
```

**Use Case:** Alert on service degradation requiring investigation.

#### Circuit Breaker Half-Open
Logged when circuit breaker enters half-open state to test recovery.

```json
{
  "level": "info",
  "event": "retry_attempt",
  "attempt": 1,
  "max_retries": 2,
  "context": "Circuit breaker entering half-open state",
  "timestamp": "2025-10-23T14:33:15.234Z"
}
```

#### Circuit Breaker Closed
Logged when circuit breaker closes after successful recovery.

```json
{
  "level": "info",
  "event": "retry_attempt",
  "attempt": 2,
  "max_retries": 2,
  "context": "Circuit breaker closed - service recovered",
  "timestamp": "2025-10-23T14:33:17.456Z"
}
```

## Error Reasons

The `reason` field categorizes errors for easier filtering:

| Reason | Description | HTTP Status | Retryable |
|--------|-------------|-------------|-----------|
| `RATE_LIMITED` | Rate limit exceeded | 429 | Yes |
| `SERVICE_UNAVAILABLE` | Service temporarily down | 503 | Yes |
| `BAD_GATEWAY` | Gateway error | 502 | Yes |
| `GATEWAY_TIMEOUT` | Gateway timeout | 504 | Yes |
| `INTERNAL_SERVER_ERROR` | Server error | 500 | Yes |
| `REQUEST_TIMEOUT` | Request timed out | 408 | Yes |
| `CONNECTION_RESET` | Network connection reset | - | Yes |
| `TIMEOUT` | Network timeout | - | Yes |
| `DNS_ERROR` | DNS resolution failed | - | Yes |
| `DNS_LOOKUP_FAILED` | DNS lookup failed | - | Yes |
| `CONFLICT` | Concurrent modification | 409 | Yes |
| `CIRCUIT_OPEN` | Circuit breaker is open | - | No |
| `CIRCUIT_OPENED` | Circuit breaker just opened | - | No |
| `UNKNOWN` | Unknown error | - | Varies |

## Practical Examples

### Example 1: Monitor Rate Limiting

```bash
# Run command and extract rate limit events
notion-cli db query DS_ID --json --verbose 2>&1 >/dev/null | \
  jq 'select(.event == "rate_limited")'
```

Output:
```json
{"level":"warn","event":"rate_limited","attempt":1,"max_retries":3,"reason":"RATE_LIMITED","retry_after_ms":1200,"url":"/v1/databases/abc123/query","status_code":429,"timestamp":"2025-10-23T14:32:15.234Z"}
```

### Example 2: Calculate Cache Hit Rate

```bash
# Run command and calculate hit rate
notion-cli search -q "test" --json --verbose 2>cache.log >/dev/null

# Count hits and misses
hits=$(jq -r 'select(.event == "cache_hit")' cache.log | wc -l)
misses=$(jq -r 'select(.event == "cache_miss")' cache.log | wc -l)
total=$((hits + misses))

echo "Cache hit rate: $((hits * 100 / total))%"
```

### Example 3: Track Retry Patterns

```bash
# Aggregate retry reasons
notion-cli db query DS_ID --json --verbose 2>&1 >/dev/null | \
  jq -r 'select(.event == "retry") | .reason' | \
  sort | uniq -c | sort -rn
```

Output:
```
  3 RATE_LIMITED
  2 SERVICE_UNAVAILABLE
  1 TIMEOUT
```

### Example 4: Alert on Circuit Breaker

```bash
# Monitor for circuit breaker events
notion-cli page retrieve PAGE_ID --json --verbose 2>&1 >/dev/null | \
  jq 'select(.reason == "CIRCUIT_OPENED" or .reason == "CIRCUIT_OPEN")'
```

### Example 5: Performance Analysis

```bash
# Track retry delays
notion-cli db query DS_ID --json --verbose 2>retries.log >/dev/null

# Calculate total retry time
jq -r 'select(.event == "retry") | .retry_after_ms' retries.log | \
  awk '{sum+=$1} END {print "Total retry delay:", sum/1000, "seconds"}'
```

### Example 6: Debug Cache Behavior

```bash
# See what's being cached
notion-cli search -q "test" --json --verbose 2>&1 >/dev/null | \
  jq 'select(.event == "cache_set") | {namespace, key, ttl_ms}'
```

Output:
```json
{"namespace":"dataSource","key":"abc123","ttl_ms":600000}
{"namespace":"page","key":"xyz789","ttl_ms":60000}
```

## Log Aggregation

### Datadog

```bash
# Send verbose logs to Datadog
notion-cli db query DS_ID --json --verbose 2>&1 >/dev/null | \
  while read line; do
    curl -X POST "https://http-intake.logs.datadoghq.com/v1/input" \
      -H "Content-Type: application/json" \
      -H "DD-API-KEY: ${DD_API_KEY}" \
      -d "$line"
  done
```

### CloudWatch Logs

```bash
# Send verbose logs to CloudWatch
notion-cli db query DS_ID --json --verbose 2>&1 >/dev/null | \
  while read line; do
    aws logs put-log-events \
      --log-group-name "/notion-cli/verbose" \
      --log-stream-name "$(date +%Y-%m-%d)" \
      --log-events timestamp=$(date +%s000),message="$line"
  done
```

### Splunk

```bash
# Send verbose logs to Splunk HTTP Event Collector
notion-cli db query DS_ID --json --verbose 2>&1 >/dev/null | \
  while read line; do
    curl -k "https://splunk.example.com:8088/services/collector" \
      -H "Authorization: Splunk ${SPLUNK_TOKEN}" \
      -d "{\"event\": $line}"
  done
```

## Best Practices

### 1. Always Separate stdout and stderr

```bash
# GOOD: Redirect stderr to file, stdout to variable
output=$(notion-cli page retrieve PAGE_ID --json --verbose 2>debug.log)

# BAD: Mix stdout and stderr
output=$(notion-cli page retrieve PAGE_ID --json --verbose 2>&1)  # Corrupts JSON
```

### 2. Use jq for Filtering

```bash
# Filter specific event types
notion-cli db query DS_ID --verbose 2>&1 >/dev/null | \
  jq 'select(.level == "error")'

# Extract specific fields
notion-cli db query DS_ID --verbose 2>&1 >/dev/null | \
  jq '{event, reason, timestamp}'
```

### 3. Monitor in Production

```bash
# Production script with error handling
#!/bin/bash
output=$(notion-cli db query DS_ID --json --verbose 2>error.log)
exit_code=$?

if [ $exit_code -ne 0 ]; then
  # Check for rate limiting
  if grep -q "RATE_LIMITED" error.log; then
    echo "Rate limited - backing off"
    sleep 60
  fi

  # Check for circuit breaker
  if grep -q "CIRCUIT_OPENED" error.log; then
    echo "Circuit breaker opened - service degraded"
    # Send alert
  fi
fi

echo "$output" | jq .
```

### 4. Disable in Production (Default)

Verbose logging is **disabled by default** to minimize overhead. Only enable when debugging or monitoring specific issues.

```bash
# Production: No verbose logging
notion-cli db query DS_ID --json

# Debug: Enable verbose logging
notion-cli db query DS_ID --json --verbose 2>debug.log
```

## Performance Impact

- **Disabled (default):** Zero overhead - logging checks are optimized away
- **Enabled:** Minimal overhead (~1-2ms per event) - JSON serialization only
- **Stderr I/O:** Asynchronous - does not block API calls

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NOTION_CLI_VERBOSE` | Enable verbose logging | `false` |
| `NOTION_CLI_DEBUG` | Enable debug mode (includes verbose) | `false` |
| `DEBUG` | Enable all debug features | `false` |

### Retry Configuration (affects retry events)

| Variable | Description | Default |
|----------|-------------|---------|
| `NOTION_CLI_MAX_RETRIES` | Max retry attempts | `3` |
| `NOTION_CLI_BASE_DELAY` | Base retry delay (ms) | `1000` |
| `NOTION_CLI_MAX_DELAY` | Max retry delay (ms) | `30000` |

### Cache Configuration (affects cache events)

| Variable | Description | Default |
|----------|-------------|---------|
| `NOTION_CLI_CACHE_ENABLED` | Enable caching | `true` |
| `NOTION_CLI_CACHE_MAX_SIZE` | Max cache entries | `1000` |
| `NOTION_CLI_CACHE_DS_TTL` | Data source TTL (ms) | `600000` (10 min) |
| `NOTION_CLI_CACHE_PAGE_TTL` | Page TTL (ms) | `60000` (1 min) |

## Troubleshooting

### Issue: No verbose logs appearing

**Solution:** Ensure verbose mode is enabled:
```bash
notion-cli page retrieve PAGE_ID --json --verbose 2>debug.log
# OR
export NOTION_CLI_VERBOSE=true
```

### Issue: Verbose logs mixing with JSON output

**Problem:** Using `2>&1` redirects stderr to stdout.

**Solution:** Keep streams separate:
```bash
# GOOD
notion-cli page retrieve PAGE_ID --json --verbose 2>debug.log >output.json

# BAD
notion-cli page retrieve PAGE_ID --json --verbose 2>&1 >output.json
```

### Issue: Too many cache events

**Solution:** Cache events are logged at `debug` level. They're verbose by design. Filter them out if needed:
```bash
notion-cli db query DS_ID --verbose 2>&1 >/dev/null | \
  jq 'select(.event != "cache_hit" and .event != "cache_miss")'
```

### Issue: Want to log only errors

**Solution:** Filter by level:
```bash
notion-cli db query DS_ID --verbose 2>&1 >/dev/null | \
  jq 'select(.level == "error" or .level == "warn")'
```

## Event Schema Reference

### Retry Event Schema

```typescript
interface RetryEvent {
  level: 'info' | 'warn' | 'error'
  event: 'retry' | 'retry_attempt' | 'retry_exhausted' | 'rate_limited'
  attempt: number
  max_retries: number
  reason?: 'RATE_LIMITED' | 'SERVICE_UNAVAILABLE' | 'TIMEOUT' | ...
  retry_after_ms?: number
  url?: string
  context?: string
  status_code?: number
  error_code?: string
  timestamp: string  // ISO 8601
}
```

### Cache Event Schema

```typescript
interface CacheEvent {
  level: 'debug' | 'info'
  event: 'cache_hit' | 'cache_miss' | 'cache_set' | 'cache_invalidate' | 'cache_evict'
  namespace: 'dataSource' | 'database' | 'user' | 'page' | 'block'
  key?: string
  age_ms?: number
  ttl_ms?: number
  cache_size?: number
  timestamp: string  // ISO 8601
}
```

## Related Documentation

- [ENHANCEMENTS.md](../ENHANCEMENTS.md) - Caching and retry features
- [OUTPUT_FORMATS.md](../OUTPUT_FORMATS.md) - Output format reference
- [README.md](../README.md) - Main documentation
- [CLAUDE.md](../CLAUDE.md) - Development guide

## Support

For issues or questions about verbose logging:
1. Check this documentation
2. Review event schemas above
3. Test with `--verbose` flag and `jq` for debugging
4. Open an issue with verbose logs attached

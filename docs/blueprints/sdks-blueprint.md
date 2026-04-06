# Heimdall SDKs — Blueprint

## What they are

The Heimdall SDKs are lightweight client libraries that let any application send structured log entries directly to a Heimdall instance — no platform log drain, no syslog daemon, no OpenTelemetry collector required. You add the package, call `monitor.error(...)`, and logs flow immediately into Heimdall's classification and agent pipeline.

Three SDKs ship with Heimdall, covering the most common server-side runtimes:

| SDK | Package name | Registry | Language | Min version |
|-----|-------------|----------|----------|-------------|
| JS/TS | `@heimdall/sdk` | npm | TypeScript | Node.js 18+ / Bun / Deno / Workers |
| Python | `heimdall-sdk` | PyPI | Python | 3.9+ |
| Go | `github.com/hejijunhao/heimdall/sdk-go` | pkg.go.dev | Go | 1.22+ |

All three are at **v0.1.0** and have zero runtime dependencies.

---

## Where the SDKs fit in the ingestion picture

Heimdall supports five distinct ingestion paths. The SDKs are one of them:

```
┌──────────────────────────────────────────────────────────────────┐
│                          HEIMDALL                                │
│                                                                  │
│  ┌──────────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │   Webhook /  │  │  OTLP    │  │  Syslog  │  │  API     │    │
│  │   SDK        │  │  HTTP    │  │  TCP/TLS │  │  Pollers │    │
│  │ (this doc)   │  │ receiver │  │ listener │  │ (Fly/    │    │
│  └──────┬───────┘  └────┬─────┘  └────┬─────┘  │  Vercel/ │    │
│         │               │              │         │  Railway/ │    │
│         └───────────────┴──────────────┘         │  MongoDB)│    │
│                         │                         └────┬─────┘    │
│                         └──────────────────────────────┘          │
│                                      │                            │
│                         ┌────────────▼───────────┐               │
│                         │   Lumber classifier    │               │
│                         │   (ONNX, threshold 0.5)│               │
│                         └────────────┬───────────┘               │
│                                      │ escalate flagged           │
│                         ┌────────────▼───────────┐               │
│                         │  Claude agent loop     │               │
│                         │  (investigate, report) │               │
│                         └────────────────────────┘               │
└──────────────────────────────────────────────────────────────────┘

         ▲
         │  POST /api/webhooks/logs
         │  Authorization: Bearer <token>
         │
  ┌──────┴──────────────────────────────┐
  │  @heimdall/sdk  /  heimdall-sdk  /  │
  │  github.com/.../sdk-go              │
  └──────┬──────────────────────────────┘
         │
  Application code (web server, worker, script, Lambda)
```

The SDKs post to **`POST /api/webhooks/logs`** — the same endpoint used by platform log drains. This means:

- A `webhook_logs` connection must be created in the Heimdall UI first (which generates the bearer token)
- Log entries pass through the identical pipeline as all other sources: Lumber classification → agent escalation → investigation → reports
- No dedicated backend endpoint or connection type is needed for SDK-based ingestion

---

## Who uses them and why

**Primary audience:** Developers who own the application code and want structured, application-level observability — not just infrastructure logs.

**Why not use a platform log drain instead?**

| Scenario | Best fit |
|----------|----------|
| You control the app and want to emit semantic events (`user.signup`, `payment.failed`) | **SDK** — you own the payload shape |
| Your platform supports log drains (Fly.io, Vercel, Railway) | Platform log drain / API poller |
| Your infrastructure can send syslog | Syslog listener |
| Your stack already uses OpenTelemetry | OTLP HTTP receiver |
| You want structured context alongside the log message | **SDK** — payload is arbitrary JSON |

The SDK is the right choice when you want to deliberately instrument your application: you know what events matter to Heimdall, and you want to shape the payload rather than relying on log parsing.

**Typical callers:**
- Backend API servers (Node/Express, FastAPI/Django, Go HTTP servers) logging business events
- Background workers and queue consumers logging job outcomes
- Serverless functions (Lambda, Cloudflare Workers) where syslog is not available
- Scripts and CLI tools that need lightweight observability without a full logging framework

---

## Shared design across all three SDKs

All three SDKs implement an identical architecture and wire protocol. The differences are language idioms only.

### Batching

Log calls are synchronous and non-blocking. Entries accumulate in an in-memory buffer. A flush is triggered when **either** condition is met:

1. **Buffer fills to `batchSize`** (default: 25) — high-throughput apps flush frequently and efficiently
2. **Timer fires after `flushInterval`** (default: 5 s) — low-throughput apps don't lose entries sitting in a buffer that never fills

On flush, the entire buffer is sent in a single HTTP POST as a JSON array. The timer resets after each flush and only starts when the buffer is non-empty.

### Retry policy

| HTTP response | Behaviour |
|---------------|-----------|
| 2xx | Success |
| 4xx | Permanent failure — calls `onError`, no retry (bad token or bad payload won't be fixed by retrying) |
| 5xx | Transient failure — retries up to `maxRetries` (default: 3) with exponential backoff |
| Network error | Same as 5xx — retries with backoff |

Backoff formula: `retryDelay × 2^attempt`. With defaults: 1 s → 2 s → 4 s. After exhausting all retries, calls `onError(error, entries)` so the application can log the failure or persist the dropped batch.

### Graceful shutdown

Each SDK has a `shutdown()` method (or `Shutdown()` in Go). This method:

1. Cancels the pending flush timer
2. Flushes the remaining buffer
3. **Waits for all in-flight send operations to complete** before returning

This last point is critical for process exit safety — without it, in-flight HTTP requests would be abandoned. Each SDK solves this differently given its concurrency model:

| SDK | In-flight tracking mechanism |
|-----|------------------------------|
| JS | `inflightSends` Set of Promises — `shutdown()` awaits `Promise.all([...inflightSends])` |
| Python | `_send_threads` list of non-daemon `threading.Thread`s — `shutdown()` joins each with 30 s timeout |
| Go | `sync.WaitGroup` — `Shutdown()` calls `wg.Wait()` after flushing |

Call `shutdown()` on `SIGTERM`, process exit, or equivalent.

### Wire format

All three SDKs post an identical JSON structure:

```http
POST /api/webhooks/logs HTTP/1.1
Authorization: Bearer whk_a3f8b2c1d4e5f6...
Content-Type: application/json

[
  {
    "source_type": "payment.failed",
    "severity": "error",
    "payload": { "orderId": "456", "error": "card declined" }
  },
  {
    "source_type": "user.signup",
    "severity": "info",
    "payload": { "userId": "789", "plan": "pro" }
  }
]
```

**`source_type`** — freeform string, developer-chosen. Appears in the Agent Log and is searchable via `search_logs`. Recommended convention: `domain.event` (e.g. `user.signup`, `cache.miss`, `db.query.slow`).

**`severity`** — one of: `debug`, `info`, `warning`, `error`, `critical`. Maps to Lumber classifier input.

**`payload`** — arbitrary JSON object. No schema enforced. This is the structured context Heimdall's agent uses when investigating.

---

## SDK reference

### JavaScript / TypeScript — `@heimdall/sdk`

**Source:** `packages/sdk-js/`
**Published to:** npm as `@heimdall/sdk`
**Build:** `tsup` → dual ESM (`dist/index.js`) + CJS (`dist/index.cjs`) + type declarations (`dist/index.d.ts`)
**Runtime deps:** Zero — uses native `fetch` (Node 18+, Bun, Deno, Cloudflare Workers, browsers)

```typescript
import { Heimdall } from '@heimdall/sdk'

const monitor = new Heimdall({
  endpoint: 'https://heimdall.example.com',
  token: 'whk_...',

  // Optional tuning:
  batchSize: 25,       // entries before auto-flush
  flushInterval: 5000, // ms before timer-flush
  maxRetries: 3,
  retryDelay: 1000,    // base backoff (ms)
  onError: (err, entries) => console.error('Heimdall flush failed:', err.message),
})

monitor.debug('cache.miss', { key: 'user:123' })
monitor.info('user.signup', { userId: '123', plan: 'pro' })
monitor.warn('rate.limit.near', { current: 95, max: 100 })
monitor.error('payment.failed', { orderId: '456', error: 'card declined' })
monitor.critical('db.connection.lost', { host: 'prod-db' })

// Manual flush (e.g. before a Lambda invocation ends)
await monitor.flush()

// Graceful shutdown (call on SIGTERM)
await monitor.shutdown()

// Inspect buffer depth
console.log(monitor.pending)
```

**API surface:**

```typescript
class Heimdall {
  constructor(options: HeimdallOptions)
  log(severity: Severity, sourceType: string, payload?: Record<string, unknown>): void
  debug(sourceType: string, payload?: Record<string, unknown>): void
  info(sourceType: string, payload?: Record<string, unknown>): void
  warn(sourceType: string, payload?: Record<string, unknown>): void
  error(sourceType: string, payload?: Record<string, unknown>): void
  critical(sourceType: string, payload?: Record<string, unknown>): void
  flush(): Promise<void>
  shutdown(): Promise<void>
  get pending(): number
}
```

---

### Python — `heimdall-sdk`

**Source:** `packages/sdk-python/`
**Published to:** PyPI as `heimdall-sdk`
**Build:** hatchling (PEP 517)
**Runtime deps:** Zero — uses stdlib `urllib.request` (no `requests` required)
**Supports:** Python 3.9, 3.10, 3.11, 3.12, 3.13

```python
from heimdall_sdk import Heimdall, HeimdallOptions

monitor = Heimdall(HeimdallOptions(
    endpoint="https://heimdall.example.com",
    token="whk_...",

    # Optional tuning:
    batch_size=25,
    flush_interval=5.0,   # seconds
    max_retries=3,
    retry_delay=1.0,      # base backoff (seconds)
    on_error=lambda err, entries: print(f"Heimdall flush failed: {err}"),
))

monitor.debug("cache.miss", {"key": "user:123"})
monitor.info("user.signup", {"user_id": "123", "plan": "pro"})
monitor.warn("rate.limit.near", {"current": 95, "max": 100})
monitor.error("payment.failed", {"order_id": "456", "error": "card declined"})
monitor.critical("db.connection.lost", {"host": "prod-db"})

# Manual flush
monitor.flush()

# Graceful shutdown (call on SIGTERM / atexit)
monitor.shutdown()

# Inspect buffer depth
print(monitor.pending)
```

**API surface:**

```python
class Heimdall:
    def __init__(self, options: HeimdallOptions) -> None
    def log(self, severity: str, source_type: str, payload: dict | None = None) -> None
    def debug(self, source_type: str, payload: dict | None = None) -> None
    def info(self, source_type: str, payload: dict | None = None) -> None
    def warn(self, source_type: str, payload: dict | None = None) -> None
    def error(self, source_type: str, payload: dict | None = None) -> None
    def critical(self, source_type: str, payload: dict | None = None) -> None
    def flush(self) -> None
    def shutdown(self) -> None  # blocks until all in-flight sends complete
    @property
    def pending(self) -> int
```

**Threading model:** Each flush spawns a non-daemon `threading.Thread` for the HTTP send. Non-daemon threads prevent the process from exiting mid-send. `shutdown()` joins all in-flight threads (30 s timeout each).

---

### Go — `github.com/hejijunhao/heimdall/sdk-go`

**Source:** `packages/sdk-go/`
**Published to:** pkg.go.dev (consumed via `go get github.com/hejijunhao/heimdall/sdk-go`)
**Runtime deps:** Zero — uses stdlib `net/http`
**Requires:** Go 1.22+

```go
import "github.com/hejijunhao/heimdall/sdk-go"

monitor := heimdall.New(heimdall.Options{
    Endpoint: "https://heimdall.example.com",
    Token:    "whk_...",

    // Optional tuning:
    BatchSize:     25,
    FlushInterval: 5 * time.Second,
    MaxRetries:    3,
    RetryDelay:    time.Second,
    OnError: func(err error, entries []heimdall.LogEntry) {
        log.Printf("Heimdall flush failed: %v", err)
    },
    HTTPClient: nil, // uses default 30 s timeout client
})
defer monitor.Shutdown()

monitor.Debug("cache.miss", map[string]any{"key": "user:123"})
monitor.Info("user.signup", map[string]any{"user_id": "123", "plan": "pro"})
monitor.Warn("rate.limit.near", map[string]any{"current": 95, "max": 100})
monitor.Error("payment.failed", map[string]any{"order_id": "456"})
monitor.Critical("db.connection.lost", map[string]any{"host": "prod-db"})

// Manual flush
monitor.Flush()

// Graceful shutdown — blocks until WaitGroup drains
monitor.Shutdown()

// Inspect buffer depth
n := monitor.Pending()
```

**API surface:**

```go
type Options struct {
    Endpoint      string
    Token         string
    BatchSize     int
    FlushInterval time.Duration
    MaxRetries    int
    RetryDelay    time.Duration
    OnError       func(err error, entries []LogEntry)
    HTTPClient    *http.Client  // inject custom client (e.g. for testing)
}

func New(opts Options) *Client

func (c *Client) Log(severity, sourceType string, payload map[string]any)
func (c *Client) Debug(sourceType string, payload map[string]any)
func (c *Client) Info(sourceType string, payload map[string]any)
func (c *Client) Warn(sourceType string, payload map[string]any)
func (c *Client) Error(sourceType string, payload map[string]any)
func (c *Client) Critical(sourceType string, payload map[string]any)
func (c *Client) Flush()
func (c *Client) Shutdown()
func (c *Client) Pending() int
```

**Notable:** Go is the only SDK that exposes an `HTTPClient` option, making it straightforward to inject a custom transport in tests or to configure mTLS.

---

## Publishing checklist

### npm — `@heimdall/sdk`

1. `cd packages/sdk-js && npm run build` — produces `dist/`
2. Update `version` in `package.json`
3. `npm publish --access public` (requires npm account with access to `@heimdall` org)

### PyPI — `heimdall-sdk`

1. `cd packages/sdk-python && python -m build` — produces `dist/*.whl` and `dist/*.tar.gz`
2. Update `version` in `pyproject.toml`
3. `twine upload dist/*` (requires PyPI API token)

### Go — `github.com/hejijunhao/heimdall/sdk-go`

Go modules are consumed directly from the git repository. To publish a new version:

1. Tag the commit: `git tag sdk-go/v0.1.0 && git push origin sdk-go/v0.1.0`
2. `GOPROXY=proxy.golang.org go list -m github.com/hejijunhao/heimdall/sdk-go@v0.1.0` — triggers proxy indexing

Note: because the Go module lives in a subdirectory of the monorepo (`packages/sdk-go/`), the module path must reflect this. Tags must be prefixed with the subdirectory path (`sdk-go/vX.Y.Z`) to satisfy the Go module proxy.

---

## What is not in the SDKs

The SDKs are deliberately minimal. They do not:

- **Parse or reformat log messages** — payload is whatever the caller passes
- **Integrate with existing logging frameworks** (Pino, Winston, Python `logging`, Go `slog`) — this is a future enhancement
- **Add trace IDs, request IDs, or correlation context** automatically — callers include what they want in the payload
- **Buffer to disk** — in-memory only; entries lost if the process exits before `shutdown()` is called
- **Compress payloads** — sent as plain JSON
- **Support streaming** — batched HTTP only; the SDK is not a long-lived connection

Framework adapters (e.g. a `pino` transport, a Python `logging.Handler` subclass, a Go `slog.Handler`) would be valuable additions but are out of scope for v0.1.

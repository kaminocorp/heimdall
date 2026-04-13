# SDK Next Steps — Publishing & General Availability

All three SDKs (`@heimdall/sdk`, `heimdall-sdk`, `sdk-go`) are implemented, tested, and building. This document covers what remains before they can be installed by external users.

---

## 1. Registry account and namespace setup

### JavaScript — npm

The package is named `@heimdall/sdk`, which requires an npm organization scope.

1. Create an npm account at [npmjs.com/signup](https://www.npmjs.com/signup)
2. Create the `@heimdall` organization at [npmjs.com/org/create](https://www.npmjs.com/org/create) (free tier is sufficient)
   - If `@heimdall` is taken, alternatives: `@heimdall-ai`, `@heimdall-monitor`, `@heimdallhq`
   - Or publish unscoped as `heimdall-sdk` — no org setup needed
3. If using a different scope, update `packages/sdk-js/package.json` → `"name"` field

### Python — PyPI

The package is named `heimdall-sdk`.

1. Create a PyPI account at [pypi.org/account/register](https://pypi.org/account/register/)
2. Enable 2FA (required for new projects since 2024)
3. If `heimdall-sdk` is taken, alternatives: `heimdall-monitor`, `heimdall-logging`
4. If using a different name, update `pyproject.toml` → `[project] name` and `heimdall_sdk/__init__.py`

### Go — Module proxy

The module is `github.com/hejijunhao/heimdall/sdk-go`. Go modules are published via git tags — no registry account needed.

1. Ensure the repo is public on GitHub (or use `GONOSUMDB`/`GOPRIVATE` for private access)
2. Module path is already set in `packages/sdk-go/go.mod`

---

## 2. Package metadata

### JavaScript (`packages/sdk-js/package.json`)

Add before publishing:

```jsonc
{
  "author": "Kamino Corporation",
  "repository": {
    "type": "git",
    "url": "https://github.com/hejijunhao/heimdall",
    "directory": "packages/sdk-js"
  },
  "homepage": "https://github.com/hejijunhao/heimdall/tree/master/packages/sdk-js",
  "bugs": "https://github.com/hejijunhao/heimdall/issues",
  "keywords": ["heimdall", "monitoring", "logging", "observability", "sdk"]
}
```

### Python (`packages/sdk-python/pyproject.toml`)

Add under `[project]`:

```toml
authors = [{ name = "Kamino Corporation" }]

[project.urls]
Homepage = "https://github.com/hejijunhao/heimdall/tree/master/packages/sdk-python"
Repository = "https://github.com/hejijunhao/heimdall"
Issues = "https://github.com/hejijunhao/heimdall/issues"
```

### Go

No metadata needed — `go.mod` and the GitHub repo serve as the source of truth. A `doc.go` or package comment in `heimdall.go` is the Go convention (already present).

---

## 3. Write a README for each SDK

Each SDK needs its own `README.md` in its package directory. This is the primary discovery surface on npm/PyPI and the Go module proxy.

**Common structure:**

- One-line description
- Install command
- Quick start (5–8 lines of code)
- Configuration options table
- Severity methods table
- Batching and retry behaviour summary
- Shutdown instructions
- Link to full docs
- License

**Install commands:**

| SDK | Install |
|-----|---------|
| JS/TS | `npm install @heimdall/sdk` |
| Python | `pip install heimdall-sdk` |
| Go | `go get github.com/hejijunhao/heimdall/sdk-go` |

---

## 4. Add a LICENSE file to each SDK

Create `LICENSE` in each package directory — MIT license text with the current year and "Kamino Corporation" as the copyright holder.

- `packages/sdk-js/LICENSE`
- `packages/sdk-python/LICENSE`
- `packages/sdk-go/LICENSE`

The `"license": "MIT"` in `package.json` and `pyproject.toml` must match.

---

## 5. Verify builds

```bash
# JavaScript
cd packages/sdk-js
npm run build        # tsup → dist/index.js, dist/index.cjs, dist/index.d.ts
npm run test         # vitest → 10 tests
npm pack --dry-run   # should include only dist/ + package.json + README + LICENSE

# Python
cd packages/sdk-python
python3 -m build     # requires `pip install build`
python3 -m pytest tests/ -v  # 9 tests
twine check dist/*   # validate package metadata

# Go
cd packages/sdk-go
go build ./...
go test ./...        # 9 tests
go vet ./...
```

### JS `npm pack --dry-run` should include only:

- `package.json`, `README.md`, `LICENSE`
- `dist/index.js`, `dist/index.cjs`, `dist/index.d.ts`, `dist/index.d.cts`

The `"files": ["dist"]` field in `package.json` controls this.

### Python `build` output should include:

- `heimdall_sdk-0.1.0.tar.gz` (sdist)
- `heimdall_sdk-0.1.0-py3-none-any.whl` (wheel)

---

## 6. Publish

### JavaScript

```bash
npm login
cd packages/sdk-js
npm publish --access public   # --access public required for first scoped publish
```

Verify at `https://www.npmjs.com/package/@heimdall/sdk`.

### Python

```bash
pip install twine build
cd packages/sdk-python
python3 -m build
twine upload dist/*           # prompts for PyPI credentials
```

Verify at `https://pypi.org/project/heimdall-sdk/`.

### Go

Go modules are published by pushing a git tag. Because the SDK lives in a subdirectory, the tag must be prefixed with the module path:

```bash
git tag sdk-go/v0.1.0
git push origin sdk-go/v0.1.0
```

The Go module proxy (`proxy.golang.org`) picks it up automatically. Verify with:

```bash
GOPROXY=https://proxy.golang.org go list -m github.com/hejijunhao/heimdall/sdk-go@v0.1.0
```

---

## 7. CI/CD (optional but recommended)

Automate publishing on git tags to avoid manual steps.

### JavaScript — GitHub Actions

```yaml
name: Publish JS SDK
on:
  push:
    tags: ['sdk-js-v*']

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          registry-url: https://registry.npmjs.org
      - run: cd packages/sdk-js && npm ci && npm run build && npm run test
      - run: cd packages/sdk-js && npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Python — GitHub Actions

```yaml
name: Publish Python SDK
on:
  push:
    tags: ['sdk-py-v*']

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - run: pip install build twine
      - run: cd packages/sdk-python && python -m build
      - run: cd packages/sdk-python && python -m pytest tests/ -v
      - run: cd packages/sdk-python && twine upload dist/*
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
```

### Go

No CI needed for publishing — the Go module proxy pulls from git tags automatically. Just ensure tests pass before tagging:

```yaml
name: Test Go SDK
on:
  push:
    paths: ['packages/sdk-go/**']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: cd packages/sdk-go && go test ./...
```

### Secrets to configure

| Secret | Registry | How to get |
|--------|----------|-----------|
| `NPM_TOKEN` | npm | npmjs.com → Access Tokens → Generate (Automation) |
| `PYPI_TOKEN` | PyPI | pypi.org → Account Settings → API tokens |

---

## 8. Documentation on the public site

Add an SDK section to Heimdall's public docs/website:

1. **Connections page in the app** — when a user creates a `webhook_logs` connection, the wizard could show "Use with SDK" instructions with language tabs (JS / Python / Go)
2. **Public docs page** — install, configure, usage examples per language

---

## 9. Framework-specific examples (future)

Not blocking for GA, but high value for adoption:

### Express (JS)

```typescript
import { Heimdall } from '@heimdall/sdk'

const monitor = new Heimdall({ endpoint: '...', token: '...' })

app.use((req, res, next) => {
  const start = Date.now()
  res.on('finish', () => {
    monitor.log(
      res.statusCode >= 500 ? 'error' : 'info',
      'http.request',
      {
        method: req.method,
        path: req.path,
        status: res.statusCode,
        duration_ms: Date.now() - start,
      },
    )
  })
  next()
})

process.on('SIGTERM', () => monitor.shutdown())
```

### Django (Python)

```python
from heimdall_sdk import Heimdall, HeimdallOptions

monitor = Heimdall(HeimdallOptions(endpoint="...", token="..."))

class HeimdallMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        import time
        start = time.time()
        response = self.get_response(request)
        monitor.log(
            "error" if response.status_code >= 500 else "info",
            "http.request",
            {
                "method": request.method,
                "path": request.path,
                "status": response.status_code,
                "duration_ms": round((time.time() - start) * 1000),
            },
        )
        return response
```

### net/http (Go)

```go
monitor := heimdall.New(heimdall.Options{Endpoint: "...", Token: "..."})
defer monitor.Shutdown()

mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // ... handle request ...
    monitor.Info("http.request", map[string]any{
        "method":      r.Method,
        "path":        r.URL.Path,
        "duration_ms": time.Since(start).Milliseconds(),
    })
})
```

These can live as code examples in each README or as a shared `examples/` directory.

---

## 10. Versioning strategy

- **0.x.y** — pre-1.0, API may change (current phase)
- **1.0.0** — when the API surface is stable and has real-world usage
- Follow semver: patch for bug fixes, minor for new features, major for breaking changes
- Each SDK is versioned independently — the webhook API contract is the coupling point, and it's stable
- Go convention: use `v0.x.y` tags prefixed with `sdk-go/` (e.g. `sdk-go/v0.1.0`)

---

## Checklist

| # | Task | JS | Python | Go |
|---|------|----|--------|----|
| 1 | Registry account/namespace | Not started | Not started | N/A (git tags) |
| 2 | Package metadata | Not started | Not started | Done (go.mod) |
| 3 | README.md | Not started | Not started | Not started |
| 4 | LICENSE file | Not started | Not started | Not started |
| 5 | Verify build + dry-run | Not started | Not started | Not started |
| 6 | First publish | Not started | Not started | Not started |
| 7 | CI/CD automation | Not started | Not started | Not started |
| 8 | Documentation on public site | Not started | Not started | Not started |
| 9 | Framework examples | Not started | Not started | Not started |
| 10 | Version strategy decided | Done (0.x) | Done (0.x) | Done (0.x) |

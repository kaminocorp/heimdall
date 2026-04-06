# SDK Next Steps — Publishing & General Availability

The `@heimdall/sdk` package is implemented, tested, and builds to ESM/CJS. This document covers what remains before it can be installed by external users.

---

## 1. npm account and scope setup

The package is named `@heimdall/sdk`, which requires an npm organization scope.

**Steps:**

1. Create an npm account at [npmjs.com/signup](https://www.npmjs.com/signup) (or use an existing one)
2. Create the `@heimdall` organization at [npmjs.com/org/create](https://www.npmjs.com/org/create)
   - Free tier (public packages only) is sufficient
   - If `@heimdall` is taken, alternatives: `@heimdall-ai`, `@heimdall-monitor`, `@heimdallhq`
3. If using a different scope, update `package.json` → `"name"` field accordingly

**Alternative:** Publish unscoped as `heimdall-sdk` — no org setup needed, but less namespace protection.

---

## 2. Package metadata

Before publishing, update `packages/sdk-js/package.json` with:

```jsonc
{
  "name": "@heimdall/sdk",
  "version": "0.1.0",
  "description": "Lightweight JavaScript/TypeScript SDK for sending logs to Heimdall",
  "license": "MIT",
  // Add these:
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

---

## 3. Write a README

Create `packages/sdk-js/README.md` with:

- One-line description
- Install command (`npm install @heimdall/sdk`)
- Quick start (5 lines of code)
- Configuration options table
- Severity methods table
- Batching and retry behaviour summary
- Shutdown instructions
- Link to full docs
- License

This README is what npm displays on the package page — it's the primary discovery surface.

---

## 4. Add a LICENSE file

Create `packages/sdk-js/LICENSE` — MIT license text with the current year and "Kamino Corporation" as the copyright holder. The `"license": "MIT"` in package.json must match.

---

## 5. Verify the build

```bash
cd packages/sdk-js
npm run build        # tsup → dist/index.js, dist/index.cjs, dist/index.d.ts
npm run test         # vitest → 10 tests passing
npm pack --dry-run   # preview what will be published (only dist/ + package.json + README + LICENSE)
```

Check that `npm pack --dry-run` output includes only:
- `package.json`
- `README.md`
- `LICENSE`
- `dist/index.js`
- `dist/index.cjs`
- `dist/index.d.ts`
- `dist/index.d.cts`

The `"files": ["dist"]` field in package.json controls this. Source code (`src/`) is excluded from the published package.

---

## 6. Publish

```bash
# Login (one-time)
npm login

# Publish as public scoped package
cd packages/sdk-js
npm publish --access public
```

`--access public` is required for the first publish of a scoped package (npm defaults scoped packages to private).

After publishing, verify at `https://www.npmjs.com/package/@heimdall/sdk`.

---

## 7. CI/CD (optional but recommended)

Automate publishing on git tags to avoid manual `npm publish` steps.

**GitHub Actions workflow** (`packages/sdk-js/.github/workflows/publish.yml` or repo-level):

```yaml
name: Publish SDK
on:
  push:
    tags: ['sdk-v*']

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

Publish flow: bump version in `package.json` → commit → `git tag sdk-v0.1.0` → `git push --tags`.

Store the npm token as `NPM_TOKEN` in GitHub repo secrets.

---

## 8. Documentation on the public site

Add an SDK section to Heimdall's public docs/website:

1. **Connections page in the app** — when a user creates a `webhook_logs` connection, the wizard could show "Use with SDK" instructions alongside the existing webhook setup
2. **Public docs page** — install, configure, usage examples for common frameworks:
   - Express/Koa middleware that auto-logs requests
   - Next.js API route logging
   - Serverless (Lambda handler wrapper)
   - Generic try/catch error reporting

---

## 9. Framework-specific examples (future)

Not blocking for GA, but high value for adoption:

```typescript
// Express middleware example
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

// Graceful shutdown
process.on('SIGTERM', () => monitor.shutdown())
```

These can live as code examples in the README or as a separate `examples/` directory.

---

## 10. Versioning strategy

- **0.x.y** — pre-1.0, API may change (current phase)
- **1.0.0** — when the API surface is stable and has real-world usage
- Follow semver: patch for bug fixes, minor for new features, major for breaking changes
- The SDK version is independent of Heimdall's version — the webhook API contract is the coupling point, and it's stable

---

## Checklist

| # | Task | Status |
|---|------|--------|
| 1 | npm org/scope setup | Not started |
| 2 | Package metadata (author, repo, homepage) | Not started |
| 3 | README.md | Not started |
| 4 | LICENSE file | Not started |
| 5 | Verify build + dry-run | Not started |
| 6 | First publish to npm | Not started |
| 7 | CI/CD automation | Not started |
| 8 | Documentation on public site | Not started |
| 9 | Framework examples | Not started |
| 10 | Version strategy decided | Done (0.x pre-1.0) |

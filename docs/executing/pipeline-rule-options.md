# Escalation Rule Persistence — Design Options

Captures the three architectural options considered for how Heimdall's log-escalation rules are defined and managed. The current implementation follows **Option A** (hardcoded + stable rule IDs). Options B and C are recorded for future reference — they are not on the near-term roadmap.

---

## Background

The severity gate decides which classified logs are escalated to the Claude assessment loop. Rules today live in `backend/internal/agent/severity_gate.go` as a single `switch` on `(lumber.Event.Type, lumber.Event.Category)`. They form an allow-list of 10–13 branches that return `true`.

Relevant constraints:
- Rules are tightly coupled to Lumber's taxonomy (Type/Category vocabulary).
- Rule evaluation is on the hot path — every classified log is checked.
- Changes to rules today require a code deploy.
- No per-org or per-app override exists today.

---

## Option A — Hardcoded with stable rule IDs (chosen)

### Shape

`ShouldEscalate` is refactored to return `(bool, ruleID string)`. Each branch gets a stable string ID:

| Branch | Rule ID |
|--------|---------|
| `ERROR` (any category) | `error_type` |
| `PERFORMANCE` (any category) | `performance_type` |
| `REQUEST` + `server_error` | `request_server_error` |
| `REQUEST` + `slow_request` | `request_slow_request` |
| `DEPLOY` (any category) | `deploy_type` |
| `SYSTEM` + `resource_alert` | `system_resource_alert` |
| `SYSTEM` + `config_change` | `system_config_change` |
| `ACCESS` + `login_failure` | `access_login_failure` |
| `ACCESS` + `auth_failure` | `access_auth_failure` |
| `ACCESS` + `permission_change` | `access_permission_change` |
| `ACCESS` + `api_key_event` | `access_api_key_event` |
| `UNCLASSIFIED` | `unclassified_type` |
| Unknown type (default branch) | `unknown_type_default` |
| Non-escalated (negative decision) | `""` (empty) |

The rule ID flows through to `log_pipeline_events.rule_hit` for pipeline telemetry. An empty rule ID means "no rule matched, escalation denied."

### Pros

- Zero schema changes, zero config UI.
- Rules live next to the taxonomy they depend on (Lumber types).
- Compile-time exhaustiveness and trivial to reason about.
- Aligns with Heimdall's "simple by default" design principle.
- Negligible runtime cost (unchanged from the current switch).

### Cons

- Rule changes require deploys.
- No per-org / per-app customisation.
- No UI visibility into what rules exist.

### When to revisit

- A user asks to tune rules without redeploy.
- Per-org behavioural differences become a product ask.
- A fifth or sixth rule shape is added that doesn't fit the `(type, category) → bool` pattern.

---

## Option B — DB-backed rule table

### Shape

A new `escalation_rules` table:

```sql
CREATE TABLE escalation_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID REFERENCES organizations(id) ON DELETE CASCADE, -- NULL = global default
    app_id        UUID REFERENCES applications(id)  ON DELETE CASCADE, -- NULL = org or global
    name          TEXT NOT NULL,                 -- human label
    description   TEXT,
    event_type    TEXT NOT NULL,                 -- Lumber Type, or '*' for wildcard
    category      TEXT,                          -- NULL = wildcard across categories
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, app_id, event_type, category)
);
```

Precedence (evaluated in order, first match wins):

1. App-specific rule (`app_id` set, `org_id` set).
2. Org-wide rule (`org_id` set, `app_id NULL`).
3. Global default rule (both NULL).

The gate loads active rules into an in-memory cache at startup and refreshes on write. `ShouldEscalate(event, appID)` becomes a cache lookup.

### Pros

- Tune without deploy.
- Supports per-org and per-app overrides.
- CRUD via admin API or settings UI.
- Every rule hit is already auditable via `log_pipeline_events.rule_hit` (the rule's ID).

### Cons

- Cache invalidation complexity (multi-instance deploys need pub/sub or short TTL).
- Requires a rules-management UI to be genuinely useful — significant frontend scope.
- Opens "what does Heimdall do if no rules exist?" — must seed defaults during migration.
- Adds a multi-tenancy dimension (global vs org vs app) that demands thoughtful UX.
- Lumber vocabulary drift: if Lumber adds a new type, the table doesn't automatically cover it (default-escalate vs default-drop decision).

### When to choose

- You need to tune behaviour without deploys as a real product feature.
- Users ask for per-env or per-app differences (e.g. staging drops `DEPLOY` notifications).
- A customer-success / support workflow needs to disable a noisy rule mid-incident.

### Implementation sketch

1. Migration: create `escalation_rules`, seed from hardcoded defaults (with `org_id = NULL, app_id = NULL`).
2. Gate rewrite: `ShouldEscalate` loads rules via `ruleCache.Lookup(orgID, appID, eventType, category)`.
3. Admin API: CRUD endpoints behind org-admin authz.
4. UI: settings page with rule list, enable/disable toggles, and per-app override editor.
5. Cache invalidation: PostgreSQL `LISTEN/NOTIFY` on rule writes, or 30s TTL refresh if simpler.
6. Audit trail: rule changes write to an audit log so "who disabled rule X" is answerable.

---

## Option C — Expression engine (CEL / rego / DSL)

### Shape

Rules are small expressions:

```
event.type == "ERROR" || (event.type == "REQUEST" && event.category in ["server_error", "slow_request"])
```

…or a DSL like:

```
when type=ERROR then escalate
when type=REQUEST and category in [server_error, slow_request] then escalate
when type=PERFORMANCE and confidence > 0.7 then escalate
```

Stored in DB, evaluated via Google's CEL, Open Policy Agent's rego, or a bespoke parser.

### Pros

- Maximum expressiveness — rules can combine arbitrary predicates, numeric thresholds on confidence, time-of-day conditions.
- Industry-standard (CEL, rego) has tooling, linters, fuzz testing.
- Arbitrary user rules without engineering changes.

### Cons

- Large surface area — users can write slow or malformed expressions.
- Evaluation is substantially slower than a map lookup (µs vs ns).
- Debugging a misfiring rule is harder than debugging a switch branch.
- Validation, sandboxing, resource limits all need to be built.
- Security: a rule author could accidentally DoS the gate with a pathological expression.
- Massively overshoots current product need.

### When to choose

- Users explicitly ask for compound conditions beyond `(type, category)`.
- Integrations with compliance / audit frameworks demand declarative rule definitions.
- A third-party ecosystem around rule authoring becomes valuable.

### Implementation sketch

1. Pick an engine — CEL is the most mainstream for production Go services.
2. Define a binding schema (what fields the rule can reference: type, category, confidence, severity, source_type, timestamp).
3. Compile rules at load time; fail validation on parse errors.
4. Enforce per-evaluation budget (e.g. 1ms) to prevent pathological expressions.
5. Rule-authoring UI with live-preview against recent events.
6. Observability: per-rule evaluation latency histograms so slow rules are visible.

---

## Summary table

| Dimension | A (hardcoded + IDs) | B (DB table) | C (expression engine) |
|-----------|---------------------|--------------|-----------------------|
| Change without deploy | No | Yes | Yes |
| Per-org customisation | No | Yes | Yes |
| Per-app customisation | No | Yes (via `app_id`) | Yes |
| Hot-path cost | Switch (~ns) | Cache lookup (~100ns) | Expression eval (~µs) |
| Implementation cost | Small | Medium | Large |
| Debugging ease | High | Medium | Low |
| Audit / telemetry | Rule IDs in `log_pipeline_events` | Rule IDs in `log_pipeline_events` | Rule ID + expression snapshot |
| Fits current product need | Yes | Overshoots slightly | Overshoots significantly |

---

## Current decision

**Option A.** Rule IDs are added immediately for pipeline telemetry (`rule_hit` column in `log_pipeline_events`); the switch statement remains the source of truth. The `(bool, ruleID)` return signature is a structural compatibility shim — when/if Option B is chosen, no call site changes, only the gate's internals.

Option B is the natural successor if rule tuning becomes a product need. Option C is documented for completeness but not anticipated.

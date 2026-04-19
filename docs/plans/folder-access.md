## Context

Today, Heimdall asks for production-grade trust on day one. Every shipped log connector — webhook drain, syslog TLS, OTLP HTTP, Fly.io drain, GitHub App, Postgres, Supabase — requires the user to either give Heimdall live credentials to their infra, or to wire Heimdall directly into their cloud account. There's no middle option for the user who is curious but not yet ready to hand over keys to the kingdom, and no easy onramp for the smaller team that doesn't have an ops person to do the wiring at all.

**Folder access** fills that gap. The user nominates a folder they already own — an S3 bucket, a GCS bucket, a Google Drive folder, a Dropbox folder — and dumps log files into it on whatever cadence suits them (a nightly cron, an export from a vendor dashboard, a Lambda that writes ALB logs). Heimdall watches that folder, ingests new and changed files, and feeds the contents into the existing pipeline. The user keeps full control over what Heimdall sees: only what they put in the folder.

This is a third trust tier between today's two endpoints:

| Tier | Trust required | Status |
|---|---|---|
| **Full SaaS connectors** | Heimdall holds creds to your infra | ✅ shipped |
| **Folder access** | Heimdall only sees what you choose to put in a folder | ❌ proposed (this doc) |
| **Self-host** | Heimdall runs in your VPC | ❌ explicitly out of scope |

It's also a strong fit for the existing 0.46.x **source-filtering & org-level connections** substrate: one Drive folder containing logs from N services maps cleanly onto one org-level connection with per-app source selection, no new selector UX required.

---

## Shape of the problem

### Provider priority

Three providers in scope, ordered by likely market coverage:

1. **S3 / GCS (object storage)** — first, bundled. This is where logs actually live in production today: CloudWatch Logs export, ALB/CloudFront access logs, Vercel log drains, Fly.io archive, Datadog Archives, Loki shipper output, Vector/Fluent Bit sinks. Almost every cloud-hosted service writes its logs to a bucket somewhere. The "I already have a bucket of logs, point you at it" pitch covers most of the developer-tools market in one connector. S3 and GCS are deliberately one connector type internally because the API surface is near-identical (PUT/GET/LIST/notification), the parsers are identical, and only the auth model differs (IAM access keys vs. service account JSON).
2. **Google Drive** — second. Different buyer. This is the answer for the smaller team without an ops person — the founder/PM exports CSVs from a vendor dashboard, drops them into a shared folder, and the agent picks them up. Smaller TAM than buckets, but a totally different go-to-market and trust profile, so it's a real second tranche, not a strict subset.
3. **Dropbox** — third. Same use case as Drive but a smaller footprint in Heimdall's developer/early-stage-startup audience. Worth shipping eventually for completeness; not worth blocking on.

### Change detection per provider

The three providers diverge sharply on what "watch a folder" means. Polling whole folders is wasteful and rate-limited; we want event-driven where possible.

| Provider | Native change mechanism | What we use |
|---|---|---|
| **S3** | S3 Event Notifications → SNS / SQS / EventBridge / Lambda | SQS-based push for high-volume users; bucket-prefix `LIST` with cursor as zero-config fallback |
| **GCS** | Pub/Sub notifications on bucket | Same shape as S3 — Pub/Sub push when configured, prefix `LIST` cursor as fallback |
| **Google Drive** | `changes.list` with `pageToken` (cursor-paged delta query); `changes.watch` push exists but channels expire ~24h | `changes.list` polled on a 30s tick (cheap — only deltas, not full re-scans) |
| **Dropbox** | `/files/list_folder/longpoll` (genuine HTTP long-poll) | Long-poll, with `/files/list_folder/continue` to drain |

**Important nuance for S3/GCS**: push requires the user to configure a notification target on their bucket. Most users won't do that on day one. The pull fallback (60s `LIST` with prefix + cursor on `LastModified`) is the **zero-config onramp**: paste a bucket name + IAM key, done. If the user wires up SQS/Pub/Sub later, we transparently switch to push and stop polling. Push is an upsell, not a prerequisite.

### File format contract

Detection by extension first, content sniff as fallback. Three formats covered, ~90% of real-world buckets:

1. **NDJSON** (`.ndjson`, `.json`, `.jsonl`) — one JSON object per line. The de-facto modern standard: CloudWatch Logs export, Vector, Fluent Bit, Loki, every cloud archive. Each line carries `timestamp`/`level`/`message`/`service` as named fields, so source-splitting (which app does this line belong to?) is just reading a `service` or `app` field.
2. **Plain text / syslog-style lines** (`.log`, `.txt`, no extension) — `2026-04-19T12:34:56Z INFO [auth] user signed in`. Lowest common denominator. Heimdall's existing syslog parsers and log-line heuristics handle this shape.
3. **Gzipped variants of either** (`.gz` suffix). Not a third format — a transport wrapper. CloudWatch export writes `.gz`, ALB logs write `.gz`, Fly archive writes `.gz`. Must work on day one or we miss most real S3 buckets.

For ambiguous files (no extension), sniff the first non-whitespace byte of the first line: `{` → NDJSON, anything else → plain text. CSV, Parquet, custom binary formats, multi-line JSON arrays are explicit non-goals.

### Cursor model and idempotency

Files get re-uploaded, renamed, partially appended (`tail -f` style). A naive "remember last `mtime` we saw" cursor breaks on any of those.

The cursor is `(file_identity, byte_offset)`:

- **`file_identity`** — the provider's stable identifier, *not* the path. S3/GCS: `(bucket, key, etag, version_id?)`. Drive: `fileId`. Dropbox: `id`. Renames don't change identity; overwrites change the etag/revision and force a full re-read.
- **`byte_offset`** — last successfully-ingested offset within that file. Append-only files (the common case for log rotation) can be resumed without re-reading. If etag/revision changes, offset resets to 0.

Idempotency on the line side: every ingested entry gets a deterministic key derived from `(connection_id, file_identity, byte_offset_at_line_start)`. The 0.45.1 RLS-on-idempotency-table machinery is the right substrate — the new connector slots in alongside the webhook ingestion path that already uses it.

---

## Proposal

### Phasing

Three releases, each independently shippable and useful.

**Phase 1 — S3 + GCS, pull mode only**
- New connector type `object_storage` covering both providers (config field `provider: "s3" | "gcs"`).
- Auth: S3 access key + secret + region; GCS service account JSON. Stored in `connection.config` JSONB, encrypted at rest via the existing connection-config encryption path.
- `PollConnector` implementation in `backend/internal/connectors/logs/object_storage.go`, polling on a 60s tick, listing bucket contents under the configured prefix, ingesting any file whose `(etag, key)` is new or whose `byte_offset < size`.
- Wires through `insertFiltered` from day one (the 0.47.3 follow-up note about poller-based connectors silently skipping the Pipeline page Ingestion stage applies here — start clean).
- File format auto-detection (extension + sniff), gzip transparent decompression, NDJSON line splitter and plain-text line splitter share the parser interface used by the 0.24.0 webhook parsers.
- Frontend: new card in the Connection Wizard's "Object storage" lane (the 0.41.1 three-lane layout already has a slot). One form for both providers, branched on the `provider` radio. Source selector (the 0.46.0 Manage Sources surface) re-used as-is — distinct file paths or top-level prefixes become "sources."

**Phase 2 — Push mode for S3/GCS, plus Drive**
- S3: optional `sqs_queue_url` field. When set, a background goroutine consumes the queue, dispatching per-object events to the same ingestion code path as the polling listener. Pull continues as a safety net at a much longer interval (5min) to catch missed events.
- GCS: same shape with a Pub/Sub subscription.
- Google Drive connector. OAuth user-token flow (Drive doesn't have a service-account model that works for non-Workspace folders). `changes.list` on a 30s tick, `pageToken` cursor.

**Phase 3 — Dropbox**
- OAuth user-token flow.
- `/files/list_folder/longpoll` driven from a connector goroutine; `/files/list_folder/continue` for the drain.

### Connector interface

The existing `PollConnector` (`backend/internal/connectors/connector.go:31-34`) is the right base for Phase 1 pull mode. Phase 2 push mode needs a long-running goroutine, which fits the existing `StreamConnector` shape. The folder-access connector implements both:

```go
type ObjectStorage struct {
    config       ObjectStorageConfig // provider, bucket, prefix, creds, optional sqs_queue_url
    connectionID uuid.UUID
    orgID        uuid.UUID
    appID        uuid.UUID            // nullable: org-scoped vs. app-scoped, per 0.46.0
    cursor       cursorStore          // per-file (etag, byte_offset) state, persisted
}

func (o *ObjectStorage) Poll(ctx context.Context, queries *db.Queries) error {
    // 1. List bucket under prefix
    // 2. For each object: fetch cursor by (connection_id, file_identity)
    // 3. If new or appended, GET range, decompress if .gz, parse, insertFiltered
    // 4. Persist updated cursor
}

func (o *ObjectStorage) Stream(ctx context.Context, out chan<- []byte) error {
    // Phase 2: only invoked when sqs_queue_url is set
}
```

### Schema additions

One new migration. Cursor state lives in its own table rather than connection.config JSONB because it's hot-path mutable and benefits from a proper index.

```sql
-- 0NN_folder_access_cursors.up.sql
CREATE TABLE folder_access_cursors (
  connection_id  uuid NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
  file_identity  text NOT NULL,           -- provider-stable identifier
  etag           text NOT NULL,           -- forces re-read on overwrite
  byte_offset    bigint NOT NULL DEFAULT 0,
  last_seen_at   timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (connection_id, file_identity)
);

CREATE INDEX idx_fac_connection_seen
  ON folder_access_cursors (connection_id, last_seen_at);
```

RLS policy mirrors the existing connection-scoped tables (read/write only when the parent connection's org matches `current_setting('app.current_user_id')`'s org). The 0.43.0 RLS-on-system-tables precedent applies directly.

A pruner pass (re-using `agent/pruner.go`) deletes cursor rows whose `last_seen_at` is older than 90 days **and** whose file is no longer present in the most recent listing — handles the "user deleted the file from the bucket" cleanup without orphaning the cursor table.

### Source mapping

In a single bucket, distinct files (or top-level prefixes — `/auth/`, `/payments/`, `/api/`) become distinct **sources** in the 0.46.0 sense. The `connection_sources` upsert pattern in `backend/internal/api/handlers/source_filter_pipeline.go:42-50` works unchanged: as new files appear, they're upserted as available sources; the user enables which ones each app should ingest via the existing Manage Sources selector.

For NDJSON files that already carry a `service` / `app` field, we get a finer-grained option: source = the value of that field, not the file path. Configurable per-connection via a `source_field` config option (default empty = use file path / prefix).

### Failure modes and observability

- **Auth expiry** (Drive/Dropbox OAuth tokens, S3 STS sessions): connector marks itself unhealthy, surfaces in the Connection list with a "Re-authorise" CTA. No silent skipping.
- **Quota / rate limits**: exponential backoff with jitter, capped at 5min between polls. Logged as WARN, not ERROR (it's a recoverable steady state on busy buckets).
- **Malformed file**: per-line parser failures don't abort the file; the line is dropped with a counter increment. Whole-file failures (corrupt gzip, etc.) advance the cursor past the file with an error annotation so we don't loop forever.
- **Prometheus**: `heimdall_folder_access_files_processed_total{provider,status}`, `heimdall_folder_access_bytes_ingested_total{provider}`, `heimdall_folder_access_cursor_lag_seconds{connection_id}` (gauge: now − max `last_seen_at` per connection). The 0.47.2 metric posture is the model.

---

## Non-goals

- **Writing files back to the user's folder.** Strictly read-only. Drainage (`docs/executing/drainage.md`) is the outbound counterpart and lives in its own subsystem.
- **Local filesystem watch.** "Point at `/var/log/`" is functionally what the SDKs and syslog connector cover — we don't need a fourth mechanism for the same shape.
- **Other formats** (CSV, Parquet, multi-line JSON arrays, custom binary). Could be added later but explicitly excluded from the 90%-coverage Phase 1 contract.
- **Cross-bucket / cross-folder federation** in a single connection. One connection = one bucket / one folder. Multiple buckets = multiple connections.
- **Backfill of historical files older than the connector's first activation.** A new connector starts its cursor at `now()`. Backfill is an explicit user action (Phase 2 polish: "import last 7d" button) — not the default, because dumping a year of CloudWatch archive into a 48h-retention pipeline serves no one.

## Open questions

1. **Region/data-residency for S3/GCS.** Heimdall's backend runs in one region. A user with a `eu-west-2` bucket pays cross-region egress. Acceptable for MVP, worth exposing in docs; longer-term may motivate per-region workers.
2. **OAuth app branding.** Drive and Dropbox both require us to register an OAuth app with Google/Dropbox, get it through their respective verification processes (Google's is non-trivial for "drive.readonly" scope), and ship a consent screen. Not a blocker, but lead time on Google verification can run multiple weeks — start that paperwork before Phase 2 code.
3. **Per-file size cap.** A 50 GiB CloudWatch export dump in one file would happily eat all worker memory. Initial cap suggestion: 500 MiB per file, configurable, log-and-skip if exceeded with a clear UI surface explaining why.

# Fix 020 — Add index on connections.status

**Issue:** #20 from scaffolding-review.md
**Date:** 2026-02-20

## Problem

No index on `connections.status`. Listing connections by status would require a full table scan.

## Fix

Added `CREATE INDEX idx_connections_status ON connections (status)` to the `001_create_connections.up.sql` migration. The down migration (`DROP TABLE`) implicitly drops the index.

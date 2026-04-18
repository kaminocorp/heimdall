import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { classifyStaleness } from '../source'

describe('classifyStaleness', () => {
  // Pin "now" to a known instant so the relative-time boundary checks are
  // deterministic. Each test body reads `now` via Date.now() via the
  // implementation under test.
  const NOW = new Date('2026-04-18T12:00:00.000Z')

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(NOW)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns "never" when last_seen_at is null (manually-added, no traffic)', () => {
    expect(classifyStaleness(null)).toBe('never')
  })

  it('returns "active" for a timestamp under 1 hour old', () => {
    const thirtyMinAgo = new Date(NOW.getTime() - 30 * 60 * 1000).toISOString()
    expect(classifyStaleness(thirtyMinAgo)).toBe('active')
  })

  it('returns "quiet" at the 1h–24h band', () => {
    const fiveHoursAgo = new Date(NOW.getTime() - 5 * 60 * 60 * 1000).toISOString()
    expect(classifyStaleness(fiveHoursAgo)).toBe('quiet')
  })

  it('returns "stale" beyond 24 hours', () => {
    const threeDaysAgo = new Date(NOW.getTime() - 3 * 24 * 60 * 60 * 1000).toISOString()
    expect(classifyStaleness(threeDaysAgo)).toBe('stale')
  })

  it('transitions "active" → "quiet" exactly at the 1h mark', () => {
    // 1h on the nose falls into the second band (strict < hour threshold).
    const oneHourAgo = new Date(NOW.getTime() - 60 * 60 * 1000).toISOString()
    expect(classifyStaleness(oneHourAgo)).toBe('quiet')
  })

  it('transitions "quiet" → "stale" exactly at the 24h mark', () => {
    const oneDayAgo = new Date(NOW.getTime() - 24 * 60 * 60 * 1000).toISOString()
    expect(classifyStaleness(oneDayAgo)).toBe('stale')
  })

  it('returns "never" for undefined (partial response / unset field)', () => {
    // Regression guard for the Phase 6 Tier 3.3 widening. Without the
    // undefined check, new Date(undefined).getTime() → NaN, and every
    // comparison against NaN is false, so the function would silently
    // classify every undefined input as 'stale' (red dot) instead of the
    // correct 'never' (muted dot).
    expect(classifyStaleness(undefined)).toBe('never')
  })

  it('accepts a Date instance in addition to an ISO string', () => {
    const thirtyMinAgo = new Date(NOW.getTime() - 30 * 60 * 1000)
    expect(classifyStaleness(thirtyMinAgo)).toBe('active')
  })

  it('returns "never" for an unparseable date string rather than classifying as stale', () => {
    // Defensive: malformed server data (bad JSON, debugging payload) should
    // land in the muted 'never' tier, not the misleading red 'stale' tier.
    expect(classifyStaleness('definitely-not-a-date')).toBe('never')
  })
})

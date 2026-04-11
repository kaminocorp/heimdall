import type { InvestigationSchedule } from '@/types/schedule'

// Small set of canonical cron presets the modal exposes as buttons. The
// ordering here is used for the button row in ScheduleModal — keep the
// cadence monotonically increasing so scanning the row reads naturally
// from "most frequent" to "least frequent".
//
// Deliberately short. Users who want exotic schedules (Sunday 3am, every
// 11 minutes, etc.) can type a raw cron expression via the advanced toggle.
// The backend validates arbitrary expressions with robfig/cron/v3.
export interface CronPreset {
  label: string
  value: string
}

export const cronPresets: CronPreset[] = [
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Daily at 9am UTC', value: '0 9 * * *' },
]

// humanizeCron renders a cron expression back into a readable label.
// Preset matches are rendered by their label; anything else is shown
// verbatim. We intentionally do NOT pull in cronstrue here — a 20KB dep
// for rendering two dozen rows of settings would be overkill, and raw
// cron is the correct affordance for users who opted into advanced mode.
export function humanizeCron(expr: string): string {
  const preset = cronPresets.find((p) => p.value === expr)
  if (preset) return preset.label
  return expr
}

// formatInterval turns a raw seconds count into a terse "every N minutes"
// / "every N hours" / "every N days" label. Matches the tone of the cron
// labels so both can share the same spot on a schedule card.
export function formatInterval(secs: number): string {
  if (secs < 60) return `Every ${secs} seconds`
  if (secs < 3600) {
    const mins = Math.round(secs / 60)
    return mins === 1 ? 'Every minute' : `Every ${mins} minutes`
  }
  if (secs < 86400) {
    const hours = Math.round(secs / 3600)
    return hours === 1 ? 'Every hour' : `Every ${hours} hours`
  }
  const days = Math.round(secs / 86400)
  return days === 1 ? 'Every day' : `Every ${days} days`
}

// formatSchedule is the single entry point callers (ScheduleCard, lists,
// summaries) use to render a schedule's cadence regardless of which mode
// it was created in. Cron wins when set — mirrors the backend's shouldFire
// precedence exactly, so the UI never disagrees with what the scheduler
// will actually do.
export function formatSchedule(s: Pick<InvestigationSchedule, 'cron_expr' | 'interval_secs'>): string {
  if (s.cron_expr && s.cron_expr.trim() !== '') {
    return humanizeCron(s.cron_expr)
  }
  return formatInterval(s.interval_secs)
}

/**
 * Extract a human-readable error message from an axios-style rejection.
 *
 * Checks, in order:
 *   1. response.data.error  — standard Heimdall JSON error field
 *   2. response.data.message — used by connection-test endpoints
 *   3. error.message — axios network error or Supabase auth error
 *   4. fallback string
 */
export function extractApiError(e: unknown, fallback: string): string {
  if (typeof e === 'object' && e !== null) {
    const err = e as { response?: { data?: { error?: string; message?: string } }; message?: string }
    if (err.response?.data?.error) return err.response.data.error
    if (err.response?.data?.message) return err.response.data.message
    if (err.message) return err.message
  }
  return fallback
}

export interface HeimdallOptions {
  /** Heimdall backend base URL (e.g. "https://heimdall.example.com") */
  endpoint: string
  /** Webhook bearer token from your Heimdall connection */
  token: string
  /** Maximum entries to buffer before flushing (default: 25) */
  batchSize?: number
  /** Maximum milliseconds to wait before flushing (default: 5000) */
  flushInterval?: number
  /** Maximum retry attempts per flush (default: 3) */
  maxRetries?: number
  /** Base delay in ms for exponential backoff (default: 1000) */
  retryDelay?: number
  /** Called when a flush permanently fails after all retries */
  onError?: (error: Error, entries: LogEntry[]) => void
}

export interface LogEntry {
  source_type: string
  severity: string
  payload: Record<string, unknown>
}

type Severity = 'debug' | 'info' | 'warning' | 'error' | 'critical'

export class Heimdall {
  private endpoint: string
  private token: string
  private batchSize: number
  private flushInterval: number
  private maxRetries: number
  private retryDelay: number
  private onError?: (error: Error, entries: LogEntry[]) => void

  private buffer: LogEntry[] = []
  private timer: ReturnType<typeof setTimeout> | null = null
  private inflightSends = new Set<Promise<void>>()

  constructor(options: HeimdallOptions) {
    this.endpoint = options.endpoint.replace(/\/+$/, '')
    this.token = options.token
    this.batchSize = options.batchSize ?? 25
    this.flushInterval = options.flushInterval ?? 5000
    this.maxRetries = options.maxRetries ?? 3
    this.retryDelay = options.retryDelay ?? 1000
    this.onError = options.onError
  }

  /**
   * Log a message with the given severity.
   *
   * @example
   * heimdall.log('info', 'user.signup', { userId: '123', plan: 'pro' })
   */
  log(severity: Severity, sourceType: string, payload: Record<string, unknown> = {}): void {
    this.buffer.push({
      source_type: sourceType,
      severity,
      payload,
    })

    if (this.buffer.length >= this.batchSize) {
      this.flush()
    } else {
      this.scheduleFlush()
    }
  }

  /** Shorthand for log('debug', ...) */
  debug(sourceType: string, payload: Record<string, unknown> = {}): void {
    this.log('debug', sourceType, payload)
  }

  /** Shorthand for log('info', ...) */
  info(sourceType: string, payload: Record<string, unknown> = {}): void {
    this.log('info', sourceType, payload)
  }

  /** Shorthand for log('warning', ...) */
  warn(sourceType: string, payload: Record<string, unknown> = {}): void {
    this.log('warning', sourceType, payload)
  }

  /** Shorthand for log('error', ...) */
  error(sourceType: string, payload: Record<string, unknown> = {}): void {
    this.log('error', sourceType, payload)
  }

  /** Shorthand for log('critical', ...) */
  critical(sourceType: string, payload: Record<string, unknown> = {}): void {
    this.log('critical', sourceType, payload)
  }

  /** Flush all buffered entries immediately. Returns when the flush completes. */
  async flush(): Promise<void> {
    this.clearTimer()

    if (this.buffer.length === 0) return

    const entries = this.buffer.splice(0)
    const p = this.send(entries)
    this.inflightSends.add(p)
    p.finally(() => this.inflightSends.delete(p))
    await p
  }

  /** Flush and stop the timer. Call this on process shutdown. */
  async shutdown(): Promise<void> {
    this.clearTimer()
    await this.flush()
    // Wait for any in-flight sends from fire-and-forget flush calls
    // (timer callbacks, batchSize triggers in log()).
    if (this.inflightSends.size > 0) {
      await Promise.all([...this.inflightSends])
    }
  }

  /** Returns the number of entries currently buffered. */
  get pending(): number {
    return this.buffer.length
  }

  private scheduleFlush(): void {
    if (this.timer) return
    this.timer = setTimeout(() => {
      this.timer = null
      this.flush()
    }, this.flushInterval)
  }

  private clearTimer(): void {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
  }

  private async send(entries: LogEntry[]): Promise<void> {
    const url = `${this.endpoint}/api/webhooks/logs`

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.token}`,
          },
          body: JSON.stringify(entries),
        })

        if (response.ok) return

        // 4xx errors are not retryable (bad token, bad payload).
        if (response.status >= 400 && response.status < 500) {
          const body = await response.text().catch(() => '')
          const err = new Error(`Heimdall SDK: ${response.status} ${body}`)
          this.onError?.(err, entries)
          return
        }

        // 5xx — retryable.
        if (attempt < this.maxRetries) {
          await this.sleep(this.retryDelay * Math.pow(2, attempt))
          continue
        }

        const err = new Error(`Heimdall SDK: ${response.status} after ${this.maxRetries + 1} attempts`)
        this.onError?.(err, entries)
        return
      } catch (e) {
        // Network error — retryable.
        if (attempt < this.maxRetries) {
          await this.sleep(this.retryDelay * Math.pow(2, attempt))
          continue
        }

        const err = e instanceof Error ? e : new Error(String(e))
        this.onError?.(err, entries)
        return
      }
    }
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

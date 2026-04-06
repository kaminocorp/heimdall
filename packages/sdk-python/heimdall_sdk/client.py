"""Heimdall SDK client with batching and retry."""

from __future__ import annotations

import json
import threading
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from typing import Any, Callable, Optional


@dataclass
class LogEntry:
    """A single log entry to be sent to Heimdall."""

    source_type: str
    severity: str
    payload: dict[str, Any]

    def to_dict(self) -> dict[str, Any]:
        return {
            "source_type": self.source_type,
            "severity": self.severity,
            "payload": self.payload,
        }


@dataclass
class HeimdallOptions:
    """Configuration options for the Heimdall client."""

    endpoint: str
    token: str
    batch_size: int = 25
    flush_interval: float = 5.0  # seconds
    max_retries: int = 3
    retry_delay: float = 1.0  # seconds (base for exponential backoff)
    on_error: Optional[Callable[[Exception, list[LogEntry]], None]] = None


class Heimdall:
    """Heimdall logging client with automatic batching and retry.

    Usage::

        from heimdall_sdk import Heimdall, HeimdallOptions

        monitor = Heimdall(HeimdallOptions(
            endpoint="https://heimdall.example.com",
            token="whk_...",
        ))

        monitor.info("user.signup", {"user_id": "123", "plan": "pro"})
        monitor.error("payment.failed", {"order_id": "456"})

        # On shutdown:
        monitor.shutdown()
    """

    def __init__(self, options: HeimdallOptions) -> None:
        self._endpoint = options.endpoint.rstrip("/")
        self._token = options.token
        self._batch_size = options.batch_size
        self._flush_interval = options.flush_interval
        self._max_retries = options.max_retries
        self._retry_delay = options.retry_delay
        self._on_error = options.on_error

        self._buffer: list[LogEntry] = []
        self._lock = threading.Lock()
        self._timer: Optional[threading.Timer] = None
        self._closed = False
        self._send_threads: list[threading.Thread] = []

    def log(
        self,
        severity: str,
        source_type: str,
        payload: Optional[dict[str, Any]] = None,
    ) -> None:
        """Log a message with the given severity."""
        entry = LogEntry(
            source_type=source_type,
            severity=severity,
            payload=payload or {},
        )

        with self._lock:
            if self._closed:
                return
            self._buffer.append(entry)
            if len(self._buffer) >= self._batch_size:
                self._flush_locked()
            else:
                self._schedule_flush()

    def debug(self, source_type: str, payload: Optional[dict[str, Any]] = None) -> None:
        self.log("debug", source_type, payload)

    def info(self, source_type: str, payload: Optional[dict[str, Any]] = None) -> None:
        self.log("info", source_type, payload)

    def warn(self, source_type: str, payload: Optional[dict[str, Any]] = None) -> None:
        self.log("warning", source_type, payload)

    def error(self, source_type: str, payload: Optional[dict[str, Any]] = None) -> None:
        self.log("error", source_type, payload)

    def critical(self, source_type: str, payload: Optional[dict[str, Any]] = None) -> None:
        self.log("critical", source_type, payload)

    def flush(self) -> None:
        """Flush all buffered entries immediately."""
        with self._lock:
            self._flush_locked()

    def shutdown(self) -> None:
        """Flush remaining entries and stop the timer.

        Waits for all in-flight send threads to complete before returning,
        ensuring no log entries are lost on process exit.
        """
        with self._lock:
            self._closed = True
            self._cancel_timer()
            self._flush_locked()
            threads = self._send_threads[:]
        # Join outside the lock so sends can complete.
        for t in threads:
            t.join(timeout=30)

    @property
    def pending(self) -> int:
        """Number of entries currently buffered."""
        with self._lock:
            return len(self._buffer)

    def _flush_locked(self) -> None:
        """Flush buffer. Must be called with self._lock held."""
        self._cancel_timer()
        if not self._buffer:
            return

        entries = self._buffer[:]
        self._buffer.clear()

        # Clean up completed send threads.
        self._send_threads = [t for t in self._send_threads if t.is_alive()]

        # Release lock during network I/O.
        # Run send in a non-daemon thread so in-flight sends survive process exit.
        # shutdown() joins these threads to ensure all data is delivered.
        t = threading.Thread(target=self._send, args=(entries,))
        t.start()
        self._send_threads.append(t)

    def _schedule_flush(self) -> None:
        """Schedule a timer-based flush if not already scheduled."""
        if self._timer is not None:
            return
        self._timer = threading.Timer(self._flush_interval, self._timer_flush)
        self._timer.daemon = True
        self._timer.start()

    def _timer_flush(self) -> None:
        with self._lock:
            self._timer = None
            self._flush_locked()

    def _cancel_timer(self) -> None:
        if self._timer is not None:
            self._timer.cancel()
            self._timer = None

    def _send(self, entries: list[LogEntry]) -> None:
        """Send entries to Heimdall with retry on 5xx/network errors."""
        url = f"{self._endpoint}/api/webhooks/logs"
        body = json.dumps([e.to_dict() for e in entries]).encode("utf-8")

        for attempt in range(self._max_retries + 1):
            try:
                req = urllib.request.Request(
                    url,
                    data=body,
                    headers={
                        "Content-Type": "application/json",
                        "Authorization": f"Bearer {self._token}",
                    },
                    method="POST",
                )
                with urllib.request.urlopen(req, timeout=30) as resp:
                    if resp.status < 300:
                        return
            except urllib.error.HTTPError as e:
                # 4xx — permanent failure, don't retry.
                if 400 <= e.code < 500:
                    if self._on_error:
                        self._on_error(
                            Exception(f"Heimdall SDK: {e.code} {e.reason}"),
                            entries,
                        )
                    return
                # 5xx — retryable.
                if attempt < self._max_retries:
                    time.sleep(self._retry_delay * (2**attempt))
                    continue
                if self._on_error:
                    self._on_error(
                        Exception(
                            f"Heimdall SDK: {e.code} after {self._max_retries + 1} attempts"
                        ),
                        entries,
                    )
                return
            except Exception as e:
                # Network error — retryable.
                if attempt < self._max_retries:
                    time.sleep(self._retry_delay * (2**attempt))
                    continue
                if self._on_error:
                    self._on_error(e, entries)
                return

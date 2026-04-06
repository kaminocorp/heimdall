"""Tests for the Heimdall Python SDK."""

import json
import threading
from unittest.mock import MagicMock, patch
from urllib.error import HTTPError

from heimdall_sdk import Heimdall, HeimdallOptions, LogEntry


def test_log_entry_to_dict():
    entry = LogEntry(source_type="test", severity="info", payload={"key": "val"})
    d = entry.to_dict()
    assert d == {"source_type": "test", "severity": "info", "payload": {"key": "val"}}


def test_buffers_entries():
    sdk = Heimdall(HeimdallOptions(
        endpoint="https://example.com",
        token="tok",
        batch_size=100,
        flush_interval=60,
    ))
    sdk.info("a", {"k": "v"})
    sdk.info("b", {"k": "v"})
    assert sdk.pending == 2
    sdk.shutdown()


def test_severity_shorthands():
    entries = []

    def capture_send(self, batch):
        entries.extend(batch)

    with patch.object(Heimdall, "_send", capture_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
            batch_size=100,
        ))
        sdk.debug("d")
        sdk.info("i")
        sdk.warn("w")
        sdk.error("e")
        sdk.critical("c")
        sdk.flush()

    severities = [e.severity for e in entries]
    assert severities == ["debug", "info", "warning", "error", "critical"]


def test_flush_sends_batch():
    sent = []

    def mock_send(self, batch):
        sent.append(batch)

    with patch.object(Heimdall, "_send", mock_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
            batch_size=100,
        ))
        sdk.info("test.event", {"key": "value"})
        sdk.flush()

    assert len(sent) == 1
    assert len(sent[0]) == 1
    assert sent[0][0].source_type == "test.event"


def test_auto_flush_on_batch_size():
    sent = []

    def mock_send(self, batch):
        sent.append(batch)

    with patch.object(Heimdall, "_send", mock_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
            batch_size=3,
            flush_interval=60,
        ))
        sdk.info("a")
        sdk.info("b")
        assert len(sent) == 0
        sdk.info("c")  # triggers flush
        # Wait for background thread.
        import time
        time.sleep(0.1)
        assert len(sent) == 1
        assert len(sent[0]) == 3
        sdk.shutdown()


def test_shutdown_flushes():
    sent = []

    def mock_send(self, batch):
        sent.append(batch)

    with patch.object(Heimdall, "_send", mock_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
            batch_size=100,
        ))
        sdk.info("final")
        sdk.shutdown()

    assert len(sent) == 1


def test_empty_flush_is_noop():
    sent = []

    def mock_send(self, batch):
        sent.append(batch)

    with patch.object(Heimdall, "_send", mock_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
        ))
        sdk.flush()

    assert len(sent) == 0


def test_endpoint_trailing_slash():
    sdk = Heimdall(HeimdallOptions(
        endpoint="https://example.com/",
        token="tok",
    ))
    assert sdk._endpoint == "https://example.com"
    sdk.shutdown()


def test_default_payload():
    sent = []

    def mock_send(self, batch):
        sent.append(batch)

    with patch.object(Heimdall, "_send", mock_send):
        sdk = Heimdall(HeimdallOptions(
            endpoint="https://example.com",
            token="tok",
        ))
        sdk.info("test")
        sdk.flush()

    assert sent[0][0].payload == {}

package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestApplySourceNamePathOverride covers the Phase 4 source_name_path
// config knob end-to-end at the helper level. Each table case represents
// a realistic payload shape and the expected post-override SourceName.
func TestApplySourceNamePathOverride(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		parsed     []webhookLogRequest
		wantNames  []string // expected SourceName after override (one per entry)
	}{
		{
			name:   "no config — parser value preserved",
			config: `{}`,
			parsed: []webhookLogRequest{
				{SourceType: "firehose", SourceName: ""},
			},
			wantNames: []string{""}, // unchanged; sourceNameOf will fall back to SourceType
		},
		{
			name:   "top-level field",
			config: `{"source_name_path":"log_group"}`,
			parsed: []webhookLogRequest{
				{SourceType: "firehose", Payload: json.RawMessage(`{"log_group":"/aws/lambda/foo"}`)},
			},
			wantNames: []string{"/aws/lambda/foo"},
		},
		{
			name:   "nested dotted path",
			config: `{"source_name_path":"meta.source"}`,
			parsed: []webhookLogRequest{
				{SourceType: "webhook", Payload: json.RawMessage(`{"meta":{"source":"svc-a"}}`)},
				{SourceType: "webhook", Payload: json.RawMessage(`{"meta":{"source":"svc-b"}}`)},
			},
			wantNames: []string{"svc-a", "svc-b"},
		},
		{
			name:   "missing path — SourceName untouched",
			config: `{"source_name_path":"nonexistent.path"}`,
			parsed: []webhookLogRequest{
				{SourceType: "webhook", SourceName: "parser-set", Payload: json.RawMessage(`{"other":"field"}`)},
			},
			wantNames: []string{"parser-set"}, // parser value stays
		},
		{
			name:   "non-string leaf — SourceName untouched",
			config: `{"source_name_path":"count"}`,
			parsed: []webhookLogRequest{
				{SourceType: "webhook", Payload: json.RawMessage(`{"count":42}`)},
			},
			wantNames: []string{""}, // number isn't a valid source_name, falls through
		},
		{
			name:   "parser value overridden when path resolves",
			config: `{"source_name_path":"override.target"}`,
			parsed: []webhookLogRequest{
				{SourceType: "webhook", SourceName: "parser-guess", Payload: json.RawMessage(`{"override":{"target":"config-winner"}}`)},
			},
			wantNames: []string{"config-winner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applySourceNamePathOverride(tt.parsed, json.RawMessage(tt.config))
			for i, want := range tt.wantNames {
				assert.Equal(t, want, tt.parsed[i].SourceName, "entry %d", i)
			}
		})
	}
}

// TestReadSourceNamePath covers the config probe in isolation. Malformed
// configs should never crash the pipeline — they just yield "".
func TestReadSourceNamePath(t *testing.T) {
	cases := map[string]string{
		`{}`:                              "",
		`{"source_name_path":""}`:         "",
		`{"source_name_path":"  foo  "}`:  "foo", // trimmed
		`{"source_name_path":"a.b.c"}`:    "a.b.c",
		`not json at all`:                 "",
		``:                                "",
		`{"other_field":"value"}`:         "",
	}
	for raw, want := range cases {
		got := readSourceNamePath(json.RawMessage(raw))
		assert.Equal(t, want, got, "input: %q", raw)
	}
}

// TestExtractStringByPath tests the dotted-path getter directly — unit
// behaviour that `applySourceNamePathOverride` depends on.
func TestExtractStringByPath(t *testing.T) {
	payload := json.RawMessage(`{"a":{"b":{"c":"deep-value","d":123}},"x":"top"}`)

	cases := []struct {
		segments []string
		want     string
	}{
		{[]string{"a", "b", "c"}, "deep-value"},
		{[]string{"x"}, "top"},
		{[]string{"a", "b", "d"}, ""},        // non-string leaf
		{[]string{"a", "missing"}, ""},       // missing key
		{[]string{"a", "b", "c", "extra"}, ""}, // past-the-end key
		{nil, ""},
		{[]string{}, ""},
	}
	for _, tc := range cases {
		got := extractStringByPath(payload, tc.segments)
		assert.Equal(t, tc.want, got, "segments=%v", tc.segments)
	}
}

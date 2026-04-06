package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapOTLPSeverity_ByNumber(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{1, "debug"},
		{5, "debug"},
		{8, "debug"},
		{9, "info"},
		{12, "info"},
		{13, "warning"},
		{16, "warning"},
		{17, "error"},
		{20, "error"},
		{21, "critical"},
		{24, "critical"},
	}
	for _, tt := range tests {
		result := mapOTLPSeverity(tt.num, "")
		assert.True(t, result.Valid)
		assert.Equal(t, tt.expected, result.String, "severityNumber=%d", tt.num)
	}
}

func TestMapOTLPSeverity_ByText(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"FATAL", "critical"},
		{"ERROR", "error"},
		{"WARN", "warning"},
		{"INFO", "info"},
		{"DEBUG", "debug"},
		{"debug2", "debug"},
		{"unknown", "info"},
	}
	for _, tt := range tests {
		result := mapOTLPSeverity(0, tt.text)
		assert.True(t, result.Valid)
		assert.Equal(t, tt.expected, result.String, "severityText=%q", tt.text)
	}
}

func TestExtractServiceName(t *testing.T) {
	attrs := []otlpKeyValue{
		{Key: "host.name", Value: otlpAnyValue{StringValue: "myhost"}},
		{Key: "service.name", Value: otlpAnyValue{StringValue: "my-api"}},
	}
	assert.Equal(t, "my-api", extractServiceName(attrs))
}

func TestExtractServiceName_Missing(t *testing.T) {
	attrs := []otlpKeyValue{
		{Key: "host.name", Value: otlpAnyValue{StringValue: "myhost"}},
	}
	assert.Equal(t, "", extractServiceName(attrs))
}

func TestFlattenAttributes(t *testing.T) {
	attrs := []otlpKeyValue{
		{Key: "k1", Value: otlpAnyValue{StringValue: "v1"}},
		{Key: "k2", Value: otlpAnyValue{IntValue: "42"}},
	}
	m := flattenAttributes(attrs)
	assert.Equal(t, "v1", m["k1"])
	assert.Equal(t, "42", m["k2"])
}

func TestFlattenAttributes_Empty(t *testing.T) {
	m := flattenAttributes(nil)
	assert.Nil(t, m)
}

func TestResolveAnyValue(t *testing.T) {
	assert.Equal(t, "hello", resolveAnyValue(otlpAnyValue{StringValue: "hello"}))
	assert.Equal(t, "42", resolveAnyValue(otlpAnyValue{IntValue: "42"}))
	assert.Equal(t, "true", resolveAnyValue(otlpAnyValue{BoolValue: true}))
	assert.Equal(t, "", resolveAnyValue(otlpAnyValue{}))
}

func TestCountLogRecords(t *testing.T) {
	req := otlpExportRequest{
		ResourceLogs: []otlpResourceLogs{
			{
				ScopeLogs: []otlpScopeLogs{
					{LogRecords: make([]otlpLogRecord, 3)},
					{LogRecords: make([]otlpLogRecord, 2)},
				},
			},
			{
				ScopeLogs: []otlpScopeLogs{
					{LogRecords: make([]otlpLogRecord, 1)},
				},
			},
		},
	}
	assert.Equal(t, 6, countLogRecords(req))
}

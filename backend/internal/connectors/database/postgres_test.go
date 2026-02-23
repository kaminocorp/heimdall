package database

import (
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := json.RawMessage(`{"host":"localhost","port":5432,"database":"test","user":"postgres","password":"pass"}`)
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil Postgres connector")
	}
}

func TestNewMissingFields(t *testing.T) {
	cfg := json.RawMessage(`{"host":""}`)
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

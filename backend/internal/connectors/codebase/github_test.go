package codebase

import (
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	// Without a GitHub client, New should return an error.
	config := json.RawMessage(`{"installation_id": 12345}`)
	_, err := New(config, nil, nil)
	if err == nil {
		t.Fatal("expected error when GitHub client is nil")
	}
}

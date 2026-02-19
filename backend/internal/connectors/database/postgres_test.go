package database

import "testing"

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil Postgres connector")
	}
}

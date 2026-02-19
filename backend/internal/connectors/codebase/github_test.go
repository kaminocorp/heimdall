package codebase

import "testing"

func TestNewGitHub(t *testing.T) {
	g := NewGitHub()
	if g == nil {
		t.Fatal("expected non-nil GitHub connector")
	}
}

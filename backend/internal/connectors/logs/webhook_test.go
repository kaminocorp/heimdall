package logs

import "testing"

func TestNewWebhook(t *testing.T) {
	w := NewWebhook()
	if w == nil {
		t.Fatal("expected non-nil Webhook connector")
	}
}

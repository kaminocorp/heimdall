package agent

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hejijunhao/heimdall/backend/internal/db"
)

func classifierLogEntry(payload string) db.LogBuffer {
	return db.LogBuffer{
		ID:      uuid.New(),
		Payload: json.RawMessage(payload),
	}
}

func TestPassthroughClassifier(t *testing.T) {
	c := &PassthroughClassifier{}

	t.Run("all logs flagged", func(t *testing.T) {
		logs := []db.LogBuffer{
			classifierLogEntry(`{"message":"hello"}`),
			classifierLogEntry(`{"message":"world"}`),
		}
		flagged, safeCount := FilterFlagged(c.Classify(logs))
		assert.Equal(t, 2, len(flagged))
		assert.Equal(t, 0, safeCount)
		for _, f := range flagged {
			assert.Equal(t, "UNCLASSIFIED", f.Type)
			assert.Equal(t, "unknown", f.Severity)
		}
	})

	t.Run("empty batch", func(t *testing.T) {
		flagged, safeCount := FilterFlagged(c.Classify(nil))
		assert.Equal(t, 0, len(flagged))
		assert.Equal(t, 0, safeCount)
	})

	t.Run("close is no-op", func(t *testing.T) {
		assert.NoError(t, c.Close())
	})
}

// skipWithoutModel skips the test if the LUMBER_MODEL_DIR env var is not set
// or the model directory doesn't exist. Integration tests only.
func skipWithoutModel(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("LUMBER_MODEL_DIR")
	if dir == "" {
		t.Skip("LUMBER_MODEL_DIR not set, skipping integration test")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skipf("model directory %s does not exist, skipping integration test", dir)
	}
	return dir
}

func TestLumberClassifier_Integration(t *testing.T) {
	modelDir := skipWithoutModel(t)

	c, err := NewLumberClassifier(modelDir)
	require.NoError(t, err)
	defer c.Close()

	t.Run("empty batch", func(t *testing.T) {
		flagged, safeCount := FilterFlagged(c.Classify(nil))
		assert.Nil(t, flagged)
		assert.Equal(t, 0, safeCount)
	})

	t.Run("known error log is flagged", func(t *testing.T) {
		logs := []db.LogBuffer{
			classifierLogEntry(`{"level":"error","message":"connection refused to db-primary:5432"}`),
		}
		flagged, safeCount := FilterFlagged(c.Classify(logs))
		assert.Equal(t, 1, len(flagged))
		assert.Equal(t, 0, safeCount)
		assert.Equal(t, "ERROR", flagged[0].Type)
	})

	t.Run("known healthy request is safe", func(t *testing.T) {
		logs := []db.LogBuffer{
			classifierLogEntry(`{"message":"GET /api/health 200 OK 3ms"}`),
		}
		flagged, safeCount := FilterFlagged(c.Classify(logs))
		assert.Equal(t, 0, len(flagged))
		assert.Equal(t, 1, safeCount)
	})

	t.Run("mixed batch", func(t *testing.T) {
		logs := []db.LogBuffer{
			classifierLogEntry(`{"message":"GET /api/users 200 OK 12ms"}`),
			classifierLogEntry(`{"message":"GET /api/health 200 OK 2ms"}`),
			classifierLogEntry(`{"message":"POST /api/login 200 OK 45ms"}`),
			classifierLogEntry(`{"message":"GET /static/app.js 200 OK 1ms"}`),
			classifierLogEntry(`{"message":"GET /api/config 200 OK 8ms"}`),
			classifierLogEntry(`{"level":"error","message":"FATAL: connection refused to db-primary:5432"}`),
			classifierLogEntry(`{"level":"error","message":"RuntimeError: null pointer dereference in handler"}`),
		}
		flagged, safeCount := FilterFlagged(c.Classify(logs))
		assert.GreaterOrEqual(t, len(flagged), 2, "at least 2 errors should be flagged")
		assert.GreaterOrEqual(t, safeCount, 3, "at least 3 healthy requests should be safe")
		assert.Equal(t, len(logs), len(flagged)+safeCount, "all logs should be accounted for")
	})
}

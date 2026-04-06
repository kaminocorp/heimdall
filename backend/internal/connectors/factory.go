package connectors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/connectors/logs"
)

// StartPoller constructs a poll-based connector for the given type and starts it.
// Returns an error if the type is unrecognised or construction fails; returns nil
// (without starting anything) for non-poll types such as webhook_logs or syslog.
func StartPoller(poller *Poller, connType string, config json.RawMessage, connID, userID uuid.UUID) error {
	switch connType {
	case "supabase":
		sb, err := logs.NewSupabase(config, connID, userID)
		if err != nil {
			return fmt.Errorf("supabase: %w", err)
		}
		poller.Start(sb, connID, time.Duration(sb.ParsedConfig().PollIntervalSecs)*time.Second)
	case "flyio":
		f, err := logs.NewFlyio(config, connID, userID)
		if err != nil {
			return fmt.Errorf("flyio: %w", err)
		}
		poller.Start(f, connID, time.Duration(f.ParsedConfig().PollIntervalSecs)*time.Second)
	case "vercel":
		v, err := logs.NewVercel(config, connID, userID)
		if err != nil {
			return fmt.Errorf("vercel: %w", err)
		}
		poller.Start(v, connID, time.Duration(v.ParsedConfig().PollIntervalSecs)*time.Second)
	case "railway":
		r, err := logs.NewRailway(config, connID, userID)
		if err != nil {
			return fmt.Errorf("railway: %w", err)
		}
		poller.Start(r, connID, time.Duration(r.ParsedConfig().PollIntervalSecs)*time.Second)
	case "mongodb":
		m, err := logs.NewMongoDB(config, connID, userID)
		if err != nil {
			return fmt.Errorf("mongodb: %w", err)
		}
		poller.Start(m, connID, time.Duration(m.ParsedConfig().PollIntervalSecs)*time.Second)
	}
	return nil
}

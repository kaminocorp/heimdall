package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/hejijunhao/heimdall/backend/internal/connectors/logs"
)

// --- Input validation helpers ---

var validConnectionTypes = map[string]bool{
	"webhook_logs": true,
	"postgres":     true,
	"supabase":     true,
	"github":       true,
	"syslog":       true,
	"otlp":         true,
	"flyio":        true,
	"vercel":       true,
	"railway":      true,
	"mongodb":      true,
}

func isValidConnectionType(t string) bool  { return validConnectionTypes[t] }
func isValidDirection(d string) bool        { return d == "one_way" || d == "two_way" }
func isValidConnectionStatus(s string) bool {
	return s == "active" || s == "inactive" || s == "error" || s == "paused"
}

// validateConnectorConfig validates a connection config by instantiating the
// appropriate connector (catching bad fields before the DB insert) and injects
// server-level syslog TLS certs when needed. Returns the (possibly modified)
// config or an error whose message is safe to send to the client.
func (s *Server) validateConnectorConfig(connType string, config json.RawMessage) (json.RawMessage, error) {
	switch connType {
	case "supabase":
		if _, err := logs.NewSupabase(config, uuid.Nil, uuid.Nil, uuid.Nil); err != nil {
			return nil, fmt.Errorf("Invalid Supabase config: %v", err)
		}
	case "flyio":
		if _, err := logs.NewFlyio(config, uuid.Nil, uuid.Nil, uuid.Nil); err != nil {
			return nil, fmt.Errorf("Invalid Fly.io config: %v", err)
		}
	case "vercel":
		if _, err := logs.NewVercel(config, uuid.Nil, uuid.Nil, uuid.Nil); err != nil {
			return nil, fmt.Errorf("Invalid Vercel config: %v", err)
		}
	case "railway":
		if _, err := logs.NewRailway(config, uuid.Nil, uuid.Nil, uuid.Nil); err != nil {
			return nil, fmt.Errorf("Invalid Railway config: %v", err)
		}
	case "mongodb":
		if _, err := logs.NewMongoDB(config, uuid.Nil, uuid.Nil, uuid.Nil); err != nil {
			return nil, fmt.Errorf("Invalid MongoDB Atlas config: %v", err)
		}
	case "webhook_logs":
		if err := validateSourceNamePathConfig(config); err != nil {
			return nil, fmt.Errorf("Invalid webhook_logs config: %v", err)
		}
	case "syslog":
		if _, err := logs.NewSyslog(config, uuid.Nil, uuid.Nil, uuid.Nil, nil); err != nil {
			return nil, fmt.Errorf("Invalid syslog config: %v", err)
		}
		if s.Config.SyslogTLSCert != "" && s.Config.SyslogTLSKey != "" {
			var cfgMap map[string]interface{}
			if err := json.Unmarshal(config, &cfgMap); err == nil {
				if cfgMap == nil {
					cfgMap = make(map[string]interface{})
				}
				if _, ok := cfgMap["tls_cert"]; !ok {
					cfgMap["tls_cert"] = s.Config.SyslogTLSCert
					cfgMap["tls_key"] = s.Config.SyslogTLSKey
					updated, err := json.Marshal(cfgMap)
					if err != nil {
						return nil, fmt.Errorf("failed to marshal syslog config")
					}
					config = updated
				}
			}
		}
	}
	return config, nil
}

// validateSourceNamePathConfig rejects malformed `source_name_path` values at
// connection write time so operators see the error immediately instead of
// debugging silent drop-by-default later. The ingestion-side extractor
// (`extractStringByPath` in webhooks.go) is deliberately defensive and
// returns "" on any malformed path — that protects the hot path, but hides
// configuration mistakes. Validating here is the belt to the extractor's
// braces.
//
// Accepts: "" (unset), "field", "a.b.c".
// Rejects: leading dot, trailing dot, empty segment anywhere, whitespace in
// segments (would never match a real JSON key and is almost certainly a
// typo). Uses the same probe unmarshal as `readSourceNamePath` so the two
// stay in sync.
func validateSourceNamePathConfig(configRaw json.RawMessage) error {
	if len(configRaw) == 0 {
		return nil
	}
	var probe struct {
		SourceNamePath *string `json:"source_name_path"`
	}
	if err := json.Unmarshal(configRaw, &probe); err != nil {
		// Parent caller handles config-malformed separately; this helper
		// is only about the source_name_path field shape.
		return nil
	}
	if probe.SourceNamePath == nil {
		return nil
	}
	raw := *probe.SourceNamePath
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		// Empty string after trim: treat as unset rather than invalid so
		// users can clear the field by submitting "" or "   ".
		return nil
	}
	if trimmed != raw {
		return fmt.Errorf("source_name_path must not have leading or trailing whitespace")
	}
	if strings.HasPrefix(trimmed, ".") || strings.HasSuffix(trimmed, ".") {
		return fmt.Errorf("source_name_path must not start or end with a dot")
	}
	for _, seg := range strings.Split(trimmed, ".") {
		if seg == "" {
			return fmt.Errorf("source_name_path must not contain empty segments (consecutive dots)")
		}
		if strings.ContainsAny(seg, " \t\n\r") {
			return fmt.Errorf("source_name_path segment %q must not contain whitespace", seg)
		}
	}
	return nil
}

package handlers

import (
	"encoding/json"
	"fmt"

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
func isValidConnectionStatus(s string) bool { return s == "active" || s == "inactive" || s == "error" }

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

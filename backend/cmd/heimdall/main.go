package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hejijunhao/heimdall/backend/internal/agent"
	"github.com/hejijunhao/heimdall/backend/internal/api"
	"github.com/hejijunhao/heimdall/backend/internal/api/middleware"
	"github.com/hejijunhao/heimdall/backend/internal/config"
	"github.com/hejijunhao/heimdall/backend/internal/connectors"
	"github.com/hejijunhao/heimdall/backend/internal/connectors/logs"
	"github.com/hejijunhao/heimdall/backend/internal/db"
	"github.com/hejijunhao/heimdall/backend/internal/github"
	"github.com/hejijunhao/heimdall/backend/internal/notifications"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	jwks := middleware.NewJWKSClient(cfg.SupabaseURL)
	if err := jwks.Fetch(context.Background()); err != nil {
		slog.Error("failed to fetch Supabase JWKS", "err", err)
		os.Exit(1)
	}
	slog.Info("fetched Supabase signing keys", "endpoint", cfg.SupabaseURL)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Initialize classifier based on CLASSIFIER_MODE
	var classifier agent.Classifier
	switch cfg.ClassifierMode {
	case "on":
		c, err := agent.NewLumberClassifier(cfg.LumberModelDir)
		if err != nil {
			slog.Error("classifier required but failed to initialize", "err", err)
			os.Exit(1)
		}
		classifier = c
		slog.Info("lumber classifier initialized", "model_dir", cfg.LumberModelDir)
	case "off":
		classifier = &agent.PassthroughClassifier{}
		slog.Info("classifier disabled, all logs will be escalated")
	default: // "fallback"
		c, err := agent.NewLumberClassifier(cfg.LumberModelDir)
		if err != nil {
			slog.Warn("lumber classifier unavailable, falling back to passthrough", "err", err)
			classifier = &agent.PassthroughClassifier{}
		} else {
			classifier = c
			slog.Info("lumber classifier initialized", "model_dir", cfg.LumberModelDir)
		}
	}

	// Initialize GitHub App client (optional — nil if not configured).
	var ghClient *github.Client
	if cfg.GitHubAppID != "" {
		var err error
		ghClient, err = github.NewClient(cfg.GitHubAppID, cfg.GitHubClientID, cfg.GitHubPrivateKey)
		if err != nil {
			slog.Error("failed to initialize GitHub App client", "err", err)
			os.Exit(1)
		}
		if ghClient != nil {
			slog.Info("GitHub App configured", "app_id", cfg.GitHubAppID)
		}
	}

	queries := db.New(pool)
	notifier := notifications.NewDispatcher(queries, cfg)
	ag := agent.New(queries, cfg, classifier, notifier, ghClient)
	ag.Start(context.Background())

	poller := connectors.NewPoller(queries)
	resumePollers(context.Background(), queries, poller)

	listener := connectors.NewListenerManager()
	resumeSyslogListeners(context.Background(), queries, listener, cfg)

	router := api.NewRouter(cfg, pool, ag, jwks, ghClient, poller, listener)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	// Use an error channel instead of os.Exit so deferred cleanup always runs.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		// Normal shutdown signal.
	case err := <-serverErr:
		slog.Error("server error, shutting down", "err", err)
	}

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "err", err)
	}

	// Stop connectors with a timeout so a hung poller or listener can't block shutdown.
	connectorDone := make(chan struct{})
	go func() {
		poller.StopAll()
		listener.StopAll()
		ag.Stop()
		close(connectorDone)
	}()
	select {
	case <-connectorDone:
		slog.Info("all connectors stopped cleanly")
	case <-time.After(10 * time.Second):
		slog.Warn("connector shutdown timed out after 10s, proceeding")
	}

	if err := classifier.Close(); err != nil {
		slog.Error("classifier shutdown error", "err", err)
	}
}

// resumeSyslogListeners restarts syslog listeners for all active syslog connections.
func resumeSyslogListeners(ctx context.Context, queries *db.Queries, lm *connectors.ListenerManager, cfg *config.Config) {
	conns, err := queries.ListActiveConnectionsByType(ctx, "syslog")
	if err != nil {
		slog.Error("failed to list active syslog connections for resume", "err", err)
		return
	}

	for _, conn := range conns {
		// Inject server-level TLS cert/key if needed.
		configJSON := conn.Config
		if cfg.SyslogTLSCert != "" && cfg.SyslogTLSKey != "" {
			var cfgMap map[string]interface{}
			if err := json.Unmarshal(configJSON, &cfgMap); err == nil {
				if _, ok := cfgMap["tls_cert"]; !ok {
					cfgMap["tls_cert"] = cfg.SyslogTLSCert
					cfgMap["tls_key"] = cfg.SyslogTLSKey
					configJSON, _ = json.Marshal(cfgMap)
				}
			}
		}

		sl, err := logs.NewSyslog(configJSON, conn.ID, conn.UserID, queries)
		if err != nil {
			slog.Error("failed to create syslog listener for resume", "connection_id", conn.ID, "err", err)
			continue
		}
		if err := lm.Start(ctx, sl, conn.ID); err != nil {
			slog.Error("failed to start syslog listener for resume", "connection_id", conn.ID, "err", err)
			continue
		}
	}

	if len(conns) > 0 {
		slog.Info("resumed syslog listeners", "count", len(conns))
	}
}

// resumePollers restarts polling goroutines for all active poll-based connections.
func resumePollers(ctx context.Context, queries *db.Queries, poller *connectors.Poller) {
	pollerTypes := []struct {
		typeName string
		start    func(conn db.Connection)
	}{
		{"supabase", func(conn db.Connection) {
			sb, err := logs.NewSupabase(conn.Config, conn.ID, conn.UserID)
			if err != nil {
				slog.Error("resume: supabase init failed", "connection_id", conn.ID, "err", err)
				return
			}
			poller.Start(sb, conn.ID, time.Duration(sb.ParsedConfig().PollIntervalSecs)*time.Second)
		}},
		{"flyio", func(conn db.Connection) {
			f, err := logs.NewFlyio(conn.Config, conn.ID, conn.UserID)
			if err != nil {
				slog.Error("resume: flyio init failed", "connection_id", conn.ID, "err", err)
				return
			}
			poller.Start(f, conn.ID, time.Duration(f.ParsedConfig().PollIntervalSecs)*time.Second)
		}},
		{"vercel", func(conn db.Connection) {
			v, err := logs.NewVercel(conn.Config, conn.ID, conn.UserID)
			if err != nil {
				slog.Error("resume: vercel init failed", "connection_id", conn.ID, "err", err)
				return
			}
			poller.Start(v, conn.ID, time.Duration(v.ParsedConfig().PollIntervalSecs)*time.Second)
		}},
		{"railway", func(conn db.Connection) {
			r, err := logs.NewRailway(conn.Config, conn.ID, conn.UserID)
			if err != nil {
				slog.Error("resume: railway init failed", "connection_id", conn.ID, "err", err)
				return
			}
			poller.Start(r, conn.ID, time.Duration(r.ParsedConfig().PollIntervalSecs)*time.Second)
		}},
		{"mongodb", func(conn db.Connection) {
			m, err := logs.NewMongoDB(conn.Config, conn.ID, conn.UserID)
			if err != nil {
				slog.Error("resume: mongodb init failed", "connection_id", conn.ID, "err", err)
				return
			}
			poller.Start(m, conn.ID, time.Duration(m.ParsedConfig().PollIntervalSecs)*time.Second)
		}},
	}

	for _, pt := range pollerTypes {
		conns, err := queries.ListActiveConnectionsByType(ctx, pt.typeName)
		if err != nil {
			slog.Error("failed to list active connections for resume", "type", pt.typeName, "err", err)
			continue
		}
		for _, conn := range conns {
			pt.start(conn)
		}
		if len(conns) > 0 {
			slog.Info("resumed pollers", "type", pt.typeName, "count", len(conns))
		}
	}
}

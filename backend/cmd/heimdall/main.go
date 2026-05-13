package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	if cfg.LogFormat == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

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

	// Phase 3 of the RLS role-split rollout: physical two-pool topology
	// with explicit sizing per parent plan §3 (app ~30, cron ~5). Both
	// pools still resolve to the same `postgres` superuser today; Phase 6
	// flips DATABASE_URL → app_user and CRON_DATABASE_URL → cron_user via
	// env vars only. The pool handles, sizing, and call shape are all in
	// place now so the privilege change is the smallest possible delta.
	appPool, err := buildPool(context.Background(), cfg.DatabaseURL, 30, "app")
	if err != nil {
		slog.Error("failed to connect to app database pool", "err", err)
		os.Exit(1)
	}
	defer appPool.Close()

	cronURL := cfg.CronDatabaseURL
	if cronURL == "" {
		// Until Phase 6 ships the env-var split, dev and current-prod both
		// run with CRON_DATABASE_URL unset. Fall back loudly so the
		// transitional shape is visible in logs.
		slog.Warn("CRON_DATABASE_URL unset; falling back to DATABASE_URL until Phase 6 of the RLS role split")
		cronURL = cfg.DatabaseURL
	}
	cronPool, err := buildPool(context.Background(), cronURL, 5, "cron")
	if err != nil {
		slog.Error("failed to connect to cron database pool", "err", err)
		// os.Exit skips deferred funcs, so the appPool.Close() defer above
		// won't fire on this error path. Drain explicitly so connections
		// closed gracefully instead of being yanked at process death.
		appPool.Close()
		os.Exit(1)
	}
	defer cronPool.Close()

	// Always-on role observation: log the resolved role for each pool and
	// emit a WARN if both pools resolve to the same role. The WARN runs
	// in every environment, so a misconfigured production deploy that
	// forgot to set HEIMDALL_ENV=production cannot boot silently as a
	// shared superuser; the log line is the visible breadcrumb. The
	// hard-fail variant below stays gated on HEIMDALL_ENV=production —
	// it's the same check, but escalated to "refuse to start" once we
	// know we're in prod.
	logPoolRoles(context.Background(), appPool, cronPool)

	if cfg.Environment == "production" {
		if err := assertRoleSplit(context.Background(), appPool, cronPool); err != nil {
			slog.Error("role-split startup invariant failed", "err", err)
			os.Exit(1)
		}
	}

	pools := db.NewPools(appPool, cronPool)

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

	notifier := notifications.NewDispatcher(pools, cfg)
	// Pipeline-page plumbing: one bus per process, one writer holding
	// only the bus (queries are caller-supplied per Write* call). The
	// writer is handed to the Agent so monitor.go can emit
	// classified/gate/assessment events; it's also reachable from the
	// HTTP ingestion handlers via s.Agent.Pipeline() so the webhook
	// path can emit the ingestion-stage event post-commit through its
	// own UserQueries scope.
	pipelineBus := agent.NewPipelineBus()
	pipelineWriter := agent.NewPipelineWriter(pipelineBus)
	ag := agent.New(pools, cfg, classifier, notifier, ghClient, pipelineWriter, pipelineBus)
	ag.Start(context.Background())

	poller := connectors.NewPoller(pools)
	listener := connectors.NewListenerManager()

	if !cfg.DisableBackgroundJobs {
		resumePollers(context.Background(), pools, poller)
		resumeSyslogListeners(context.Background(), pools, listener, cfg)
	}

	router := api.NewRouter(cfg, pools, ag, jwks, ghClient, poller, listener)

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

// buildPool parses a Postgres URL, applies the supplied MaxConns ceiling,
// and opens the pool. The label is purely for log attribution.
func buildPool(ctx context.Context, url string, maxConns int32, label string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = maxConns
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	slog.Info("database pool ready", "label", label, "max_conns", maxConns)
	return pool, nil
}

// logPoolRoles prints the authenticated role of each pool at startup,
// regardless of HEIMDALL_ENV. Emits a WARN if both pools resolve to the
// same role — the silent-misconfig insurance that catches a prod deploy
// that forgot to set HEIMDALL_ENV=production. Best-effort: any query
// failure here is logged but never fatal (a real connectivity problem
// will fail the next real query and surface that way; we don't want a
// transient probe failure to block boot).
func logPoolRoles(ctx context.Context, appPool, cronPool *pgxpool.Pool) {
	var appRole, cronRole string
	if err := appPool.QueryRow(ctx, "SELECT current_user").Scan(&appRole); err != nil {
		slog.Warn("could not probe app pool role", "err", err)
		return
	}
	if err := cronPool.QueryRow(ctx, "SELECT current_user").Scan(&cronRole); err != nil {
		slog.Warn("could not probe cron pool role", "err", err)
		return
	}
	if appRole == cronRole {
		slog.Warn("DATABASE_URL and CRON_DATABASE_URL authenticate as the same role; production requires distinct app_user / cron_user (set HEIMDALL_ENV=production to make this fatal)",
			"role", appRole)
		return
	}
	slog.Info("database pool roles", "app_role", appRole, "cron_role", cronRole)
}

// assertRoleSplit is the production hard-fail variant of logPoolRoles.
// Used only when HEIMDALL_ENV=production. The check runs `SELECT current_user`
// on each pool — that's the strongest signal that the env-var flip
// actually landed (URL string comparison would miss the case where two
// distinct URLs both still resolve to `postgres`).
func assertRoleSplit(ctx context.Context, appPool, cronPool *pgxpool.Pool) error {
	var appRole, cronRole string
	if err := appPool.QueryRow(ctx, "SELECT current_user").Scan(&appRole); err != nil {
		return err
	}
	if err := cronPool.QueryRow(ctx, "SELECT current_user").Scan(&cronRole); err != nil {
		return err
	}
	if appRole == cronRole {
		return fmt.Errorf("DATABASE_URL and CRON_DATABASE_URL both authenticate as %q; production requires distinct app_user / cron_user roles", appRole)
	}
	slog.Info("role-split verified", "app_role", appRole, "cron_role", cronRole)
	return nil
}

// resumeSyslogListeners restarts syslog listeners for all active syslog connections.
func resumeSyslogListeners(ctx context.Context, pools *db.Pools, lm *connectors.ListenerManager, cfg *config.Config) {
	conns, err := pools.CronQueries().ListActiveConnectionsByType(ctx, "syslog")
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
					if marshaled, err := json.Marshal(cfgMap); err != nil {
						slog.Error("failed to marshal syslog config for resume", "connection_id", conn.ID, "err", err)
						continue
					} else {
						configJSON = marshaled
					}
				}
			}
		}

		if conn.AppID == nil {
			// Defensive skip — syslog is app-scoped by design. If somehow a
			// syslog connection landed with NULL app_id, bailing here is safer
			// than passing uuid.Nil into NewSyslog.
			slog.Warn("skipping syslog listener resume: org-scoped not supported", "connection_id", conn.ID)
			continue
		}
		sl, err := logs.NewSyslog(configJSON, conn.ID, conn.UserID, *conn.AppID, pools)
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
func resumePollers(ctx context.Context, pools *db.Pools, poller *connectors.Poller) {
	cronQ := pools.CronQueries()
	for _, typeName := range []string{"supabase", "flyio", "vercel", "railway", "mongodb"} {
		conns, err := cronQ.ListActiveConnectionsByType(ctx, typeName)
		if err != nil {
			slog.Error("failed to list active connections for resume", "type", typeName, "err", err)
			continue
		}
		for _, conn := range conns {
			if conn.AppID == nil {
				// Pollers operate on a specific app — skip org-scoped
				// connections (which are webhook_logs / otlp only anyway).
				slog.Warn("skipping poller resume: org-scoped not supported for this type",
					"type", typeName, "connection_id", conn.ID)
				continue
			}
			if err := connectors.StartPoller(poller, typeName, conn.Config, conn.ID, conn.UserID, *conn.AppID); err != nil {
				slog.Error("resume: poller init failed", "type", typeName, "connection_id", conn.ID, "err", err)
			}
		}
		if len(conns) > 0 {
			slog.Info("resumed pollers", "type", typeName, "count", len(conns))
		}
	}
}

package main

import (
	"context"
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

	router := api.NewRouter(cfg, pool, ag, jwks, ghClient)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "err", err)
	}
	ag.Stop()
	if err := classifier.Close(); err != nil {
		slog.Error("classifier shutdown error", "err", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ninosh210804/lumiera/apps/server/internal/accounting"
	"github.com/ninosh210804/lumiera/apps/server/internal/config"
	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/eventbus"
	"github.com/ninosh210804/lumiera/apps/server/internal/handler"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	if cfg.IsDev() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var pgPool *pgxpool.Pool
	if !cfg.SQLite.Mode {
		pgPool, err = connectPostgres(ctx, cfg)
		if err != nil {
			slog.Error("connecting to PostgreSQL", "error", err)
			os.Exit(1)
		}
		defer pgPool.Close()
		slog.Info("PostgreSQL connected")
	} else {
		slog.Info("running in SQLite sidecar mode", "path", cfg.SQLite.Path)
	}

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// CORS — allow all origins in dev, tighten in prod via env
	r.Use(corsMiddleware(cfg))

	// Event bus (always created; accounting subscribes when enabled)
	bus := eventbus.New()

	// Accounting module (isolated — only wired here)
	var acctSvc handler.AccountingServicer
	if cfg.Accounting.Enabled && pgPool != nil {
		q := pgdb.New(pgPool)
		acctSvc = accounting.New(q, bus)
		slog.Info("accounting module enabled")
	}

	// Mount routes
	r.Mount("/", handler.NewRouter(pgPool, cfg, acctSvc))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.Server.Port, "env", cfg.Server.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func connectPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	poolCfg.MaxConns = cfg.Database.MaxConns
	poolCfg.MinConns = cfg.Database.MinConns
	if cfg.Database.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime
	}
	if cfg.Database.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	}

	// Set app.location_id context for every acquired connection (future use)
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging PostgreSQL: %w", err)
	}

	return pool, nil
}

func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.IsDev() {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

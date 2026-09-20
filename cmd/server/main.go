package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"plos/internal/config"
	"plos/internal/core"
	"plos/internal/domain/finance"
	transportHTTP "plos/internal/transport/http"
	"plos/migrations"
	"plos/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := core.InitDB(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := core.RunMigrations(db, migrations.FS, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	repo := finance.NewSQLiteRepository(db)
	rateService := finance.NewRateService(repo, nil)
	engine := finance.NewCashflowEngine()
	exportService := finance.NewExportService(repo)

	apiServer := transportHTTP.NewServer(repo, rateService, engine, exportService)

	router, ok := apiServer.Router().(*chi.Mux)
	if !ok {
		return errors.New("failed to typecast router to chi.Mux")
	}

	// SPA Router Middleware for embedded static assets
	distSubFS, err := fs.Sub(web.DistFS, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distSubFS))
		router.NotFound(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}

			f, openErr := distSubFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if openErr != nil {
				r.URL.Path = "/"
			} else {
				_ = f.Close()
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

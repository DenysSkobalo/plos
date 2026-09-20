package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"testing"
	"time"

	"plos/internal/core"
	"plos/internal/domain/finance"
	transportHTTP "plos/internal/transport/http"
)

//go:embed testdata/migrations/*.sql
var testMigrationsFS embed.FS

func TestServer_BootAndShutdown(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_main_*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	db, err := core.InitDB(tmpFile.Name())
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := core.RunMigrations(db, testMigrationsFS, "testdata/migrations"); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	repo := finance.NewSQLiteRepository(db)
	rateService := finance.NewRateService(repo, nil)
	engine := finance.NewCashflowEngine()
	exportService := finance.NewExportService(repo)

	server := transportHTTP.NewServer(repo, rateService, engine, exportService)

	httpServer := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: server.Router(),
	}

	go func() {
		_ = httpServer.ListenAndServe()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		t.Fatalf("HTTP server shutdown failed: %v", err)
	}
}

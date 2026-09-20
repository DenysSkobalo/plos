package http

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"plos/internal/core"
	"plos/internal/domain/finance"
)

//go:embed testdata/migrations/*.sql
var testMigrationsFS embed.FS

func setupTestServer(t *testing.T) (server *Server, cleanup func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_http_*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	_ = tmpFile.Close()

	db, err := core.InitDB(tmpFile.Name())
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := core.RunMigrations(db, testMigrationsFS, "testdata/migrations"); err != nil {
		_ = db.Close()
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("RunMigrations failed: %v", err)
	}

	repo := finance.NewSQLiteRepository(db)
	rateService := finance.NewRateService(repo, nil)
	engine := finance.NewCashflowEngine()
	exportService := finance.NewExportService(repo)

	server = NewServer(repo, rateService, engine, exportService)

	cleanup = func() {
		_ = db.Close()
		_ = os.Remove(tmpFile.Name())
	}

	return server, cleanup
}

func TestHTTP_AccountAndDebtEndpoints(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	ctx := context.Background()
	_ = server.repo.AddCurrency(ctx, finance.Currency{Code: "EUR", Symbol: "€"})
	_ = server.repo.AddCurrency(ctx, finance.Currency{Code: "UAH", Symbol: "₴"})

	acc := finance.Account{
		ID:                  "acc-1",
		Name:                "Monobank Overdraft",
		Type:                finance.AccountTypeDebt,
		Currency:            "UAH",
		InitialBalanceCents: -2091841,
		CurrentBalanceCents: -2091841,
	}
	body, _ := json.Marshal(acc)

	resp, err := http.Post(ts.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("POST /api/v1/accounts failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", resp.StatusCode)
	}

	meta := finance.DebtMetadata{
		AccountID:                 "acc-1",
		Priority:                  2,
		CreditorName:              "Monobank",
		OriginalCurrency:          "UAH",
		OriginalAmountCents:       2091841,
		MinMonthlyPaymentUAHCents: 195096,
		Status:                    finance.DebtStatusActive,
	}
	metaBody, _ := json.Marshal(meta)

	resp, err = http.Post(ts.URL+"/api/v1/debts", "application/json", bytes.NewBuffer(metaBody))
	if err != nil {
		t.Fatalf("POST /api/v1/debts failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/api/v1/debts")
	if err != nil {
		t.Fatalf("GET /api/v1/debts failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var debts []finance.DebtMetadata
	_ = json.NewDecoder(resp.Body).Decode(&debts)
	_ = resp.Body.Close()

	if len(debts) != 1 || debts[0].CreditorName != "Monobank" {
		t.Errorf("unexpected debts response: %+v", debts)
	}
}

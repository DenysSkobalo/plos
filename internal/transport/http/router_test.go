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
	"time"

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

func TestHTTP_GetTransactionsEndpoint(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	ctx := context.Background()
	_ = server.repo.AddCurrency(ctx, finance.Currency{Code: "EUR", Symbol: "€"})

	// 1. Verify empty array fallback when no transactions exist
	resp, err := http.Get(ts.URL + "/api/v1/transactions?limit=50")
	if err != nil {
		t.Fatalf("GET /api/v1/transactions failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var emptyTxs []finance.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&emptyTxs); err != nil {
		_ = resp.Body.Close()
		t.Fatalf("failed to decode empty transactions response: %v", err)
	}
	_ = resp.Body.Close()

	if emptyTxs == nil {
		t.Errorf("expected non-nil empty slice `[]`, got `null` JSON response")
	}
	if len(emptyTxs) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(emptyTxs))
	}

	// 2. Insert transaction directly into storage layer via pointer &tx
	tx := finance.Transaction{
		ID:                  "tx-test-01",
		AmountCents:         25000,
		Currency:            "EUR",
		ExchangeRateApplied: 1.0,
		Description:         "Monthly Subscription",
		ExecutionDate:       time.Now().Format(time.RFC3339),
	}
	if err := server.repo.ProcessTransaction(ctx, &tx); err != nil {
		t.Fatalf("failed to seed test transaction: %v", err)
	}

	// 3. Query endpoint with limit parameter
	resp, err = http.Get(ts.URL + "/api/v1/transactions?limit=10")
	if err != nil {
		t.Fatalf("GET /api/v1/transactions failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var txs []finance.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		_ = resp.Body.Close()
		t.Fatalf("failed to decode transactions response: %v", err)
	}
	_ = resp.Body.Close()

	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}

	if txs[0].ID != "tx-test-01" || txs[0].AmountCents != 25000 || txs[0].Currency != "EUR" {
		t.Errorf("unexpected transaction data payload: %+v", txs[0])
	}
}

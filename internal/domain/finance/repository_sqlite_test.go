package finance

import (
	"context"
	"embed"
	"os"
	"testing"

	"plos/internal/core"
)

//go:embed testdata/migrations/*.sql
var testMigrationsFS embed.FS

func setupTestDB(t *testing.T) (repo *SQLiteRepository, cleanup func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_repo_*.db")
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

	repo = NewSQLiteRepository(db)
	cleanup = func() {
		_ = db.Close()
		_ = os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

func TestSQLiteRepository_CurrenciesAndAccounts(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Усі 3 бази валюти (EUR, UAH, USD) ініціалізовані міграцією
	currencies, err := repo.GetCurrencies(ctx)
	if err != nil {
		t.Fatalf("GetCurrencies failed: %v", err)
	}
	if len(currencies) != 3 {
		t.Fatalf("expected 3 seed currencies, got %d", len(currencies))
	}

	acc := Account{
		ID:             "acc-mfo-1",
		Name:           "Moneyveo",
		Type:           AccountTypeDebt,
		Currency:       "UAH",
		InitialBalance: -7721.92,
		CurrentBalance: -7721.92,
	}
	if err := repo.CreateAccount(ctx, &acc); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	debtMeta := DebtMetadata{
		AccountID:            acc.ID,
		Priority:             1,
		CreditorName:         "Moneyveo",
		OriginalCurrency:     "UAH",
		OriginalAmount:       7721.92,
		MinMonthlyPaymentUAH: 7721.92,
		Status:               DebtStatusOverdue,
	}
	if err := repo.SetDebtMetadata(ctx, &debtMeta); err != nil {
		t.Fatalf("SetDebtMetadata failed: %v", err)
	}

	debts, err := repo.GetActiveDebtsOrderedByPriority(ctx)
	if err != nil {
		t.Fatalf("GetActiveDebtsOrderedByPriority failed: %v", err)
	}
	if len(debts) != 1 {
		t.Fatalf("expected 1 active debt, got %d", len(debts))
	}
	if debts[0].Priority != 1 || debts[0].CreditorName != "Moneyveo" {
		t.Errorf("unexpected debt metadata content: %+v", debts[0])
	}
}

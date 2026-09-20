package finance

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestExportService_ExportDebtsToCSV(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_ = repo.AddCurrency(ctx, Currency{Code: "UAH", Symbol: "₴"})

	acc := Account{
		ID:             "d1",
		Name:           "Moneyveo",
		Type:           AccountTypeDebt,
		Currency:       "UAH",
		InitialBalance: -7721.92,
		CurrentBalance: -7721.92,
	}
	_ = repo.CreateAccount(ctx, &acc)

	debtMeta := DebtMetadata{
		AccountID:            "d1",
		Priority:             1,
		CreditorName:         "Moneyveo",
		OriginalCurrency:     "UAH",
		OriginalAmount:       7721.92,
		MinMonthlyPaymentUAH: 7721.92,
		Status:               DebtStatusOverdue,
	}
	_ = repo.SetDebtMetadata(ctx, &debtMeta)

	exportService := NewExportService(repo)
	var buf bytes.Buffer

	if err := exportService.ExportDebtsToCSV(ctx, &buf); err != nil {
		t.Fatalf("ExportDebtsToCSV failed: %v", err)
	}

	csvOutput := buf.String()

	if !strings.Contains(csvOutput, "Priority,Creditor,Status,Original_Currency") {
		t.Errorf("expected CSV header, got:\n%s", csvOutput)
	}

	if !strings.Contains(csvOutput, "1,Moneyveo,OVERDUE,UAH,7721.92,-7721.92,7721.92") {
		t.Errorf("expected row data for Moneyveo, got:\n%s", csvOutput)
	}
}

package finance

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
)

type ExportService struct {
	repo Repository
}

func NewExportService(repo Repository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportDebtsToCSV генерує CSV-файл із поточним станом боргів та їх пріоритетами.
func (s *ExportService) ExportDebtsToCSV(ctx context.Context, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// RFC 4180 Header
	header := []string{
		"Priority",
		"Creditor",
		"Status",
		"Original_Currency",
		"Original_Amount",
		"Current_Balance",
		"Min_Monthly_Payment_UAH",
		"Deadline",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write csv header: %w", err)
	}

	debts, err := s.repo.GetActiveDebtsOrderedByPriority(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active debts for export: %w", err)
	}

	for _, d := range debts {
		acc, err := s.repo.GetAccountByID(ctx, d.AccountID)
		currentBalance := 0.0
		if err == nil && acc != nil {
			currentBalance = acc.CurrentBalance
		}

		deadline := ""
		if d.DeadlineDate != nil {
			deadline = *d.DeadlineDate
		}

		record := []string{
			fmt.Sprintf("%d", d.Priority),
			d.CreditorName,
			string(d.Status),
			d.OriginalCurrency,
			fmt.Sprintf("%.2f", d.OriginalAmount),
			fmt.Sprintf("%.2f", currentBalance),
			fmt.Sprintf("%.2f", d.MinMonthlyPaymentUAH),
			deadline,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write csv row: %w", err)
		}
	}

	return nil
}

// ExportTransactionsToCSV генерує CSV-файл списку фінансових транзакцій.
func (s *ExportService) ExportTransactionsToCSV(ctx context.Context, w io.Writer, limit int) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"ID",
		"Execution_Date",
		"Source_Account",
		"Destination_Account",
		"Amount",
		"Currency",
		"Exchange_Rate_Applied",
		"Description",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write csv header: %w", err)
	}

	transactions, err := s.repo.GetTransactions(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to query transactions for export: %w", err)
	}

	for _, tx := range transactions {
		sourceAcc := ""
		if tx.SourceAccountID != nil {
			sourceAcc = *tx.SourceAccountID
		}
		destAcc := ""
		if tx.DestinationAccountID != nil {
			destAcc = *tx.DestinationAccountID
		}

		record := []string{
			tx.ID,
			tx.ExecutionDate,
			sourceAcc,
			destAcc,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.Currency,
			fmt.Sprintf("%.4f", tx.ExchangeRateApplied),
			tx.Description,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write transaction csv row: %w", err)
		}
	}

	return nil
}

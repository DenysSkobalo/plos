package finance

import (
	"testing"
)

func TestCashflowEngine_SimulateCascadePayoff(t *testing.T) {
	engine := NewCashflowEngine()

	buyRate := 51.0700

	incomes := []IncomeItem{
		{Name: "Salary September", AmountCents: 123000, Currency: "EUR"},
	}

	expenses := []ExpenseItem{
		{Name: "Mobile", AmountCents: 1220, Currency: "EUR"},
		{Name: "TCL Pass", AmountCents: 2050, Currency: "EUR"},
	}

	debts := []DebtPayoffState{
		{AccountID: "d1", CreditorName: "Moneyveo", Priority: 1, OriginalCurrency: "UAH", RemainingBalanceCents: 772192, MinMonthlyPaymentUAHCents: 772192, Status: DebtStatusOverdue},
		{AccountID: "d2", CreditorName: "Monobank", Priority: 2, OriginalCurrency: "UAH", RemainingBalanceCents: 2091841, MinMonthlyPaymentUAHCents: 195096, Status: DebtStatusActive},
		{AccountID: "d3", CreditorName: "BasicFit", Priority: 3, OriginalCurrency: "EUR", RemainingBalanceCents: 7497, MinMonthlyPaymentUAHCents: 7497, Status: DebtStatusActive},
	}

	projection, updatedDebts := engine.SimulateMonth("2026-09", incomes, expenses, debts, buyRate)

	if updatedDebts[0].Status != DebtStatusPaid || updatedDebts[0].RemainingBalanceCents != 0 {
		t.Errorf("expected P1 Moneyveo to be PAID, got balance %d status %s", updatedDebts[0].RemainingBalanceCents, updatedDebts[0].Status)
	}

	if updatedDebts[2].Status != DebtStatusPaid || updatedDebts[2].RemainingBalanceCents != 0 {
		t.Errorf("expected P3 BasicFit to be PAID, got balance %d status %s", updatedDebts[2].RemainingBalanceCents, updatedDebts[2].Status)
	}

	if updatedDebts[1].RemainingBalanceCents >= 2091841 {
		t.Errorf("expected P2 Monobank balance to decrease, got %d", updatedDebts[1].RemainingBalanceCents)
	}

	if projection.FreeCashflowEURCents <= 0 {
		t.Errorf("expected positive free cashflow, got %d", projection.FreeCashflowEURCents)
	}
}

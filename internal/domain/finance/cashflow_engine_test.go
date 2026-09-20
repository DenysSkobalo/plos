package finance

import (
	"testing"
)

func TestCashflowEngine_SimulateCascadePayoff(t *testing.T) {
	engine := NewCashflowEngine()

	buyRate := 51.0700 // Monobank EUR/UAH

	incomes := []IncomeItem{
		{Name: "Salary September", Amount: 1230.00, Currency: "EUR"},
	}

	expenses := []ExpenseItem{
		{Name: "Mobile", Amount: 12.20, Currency: "EUR"},
		{Name: "TCL Pass", Amount: 20.50, Currency: "EUR"},
	}

	debts := []DebtPayoffState{
		{AccountID: "d1", CreditorName: "Moneyveo", Priority: 1, OriginalCurrency: "UAH", RemainingBalance: 7721.92, MinMonthlyPaymentUAH: 7721.92, Status: DebtStatusOverdue},
		{AccountID: "d2", CreditorName: "Monobank", Priority: 2, OriginalCurrency: "UAH", RemainingBalance: 20918.41, MinMonthlyPaymentUAH: 1950.96, Status: DebtStatusActive},
		{AccountID: "d3", CreditorName: "BasicFit", Priority: 3, OriginalCurrency: "EUR", RemainingBalance: 74.97, MinMonthlyPaymentUAH: 74.97, Status: DebtStatusActive},
	}

	// Симуляція першого місяця
	projection, updatedDebts := engine.SimulateMonth("2026-09", incomes, expenses, debts, buyRate)

	// 1. Перевірка покриття P1 (Moneyveo) та P3 (BasicFit) повністю
	if updatedDebts[0].Status != DebtStatusPaid || updatedDebts[0].RemainingBalance != 0 {
		t.Errorf("expected P1 Moneyveo to be PAID, got balance %f status %s", updatedDebts[0].RemainingBalance, updatedDebts[0].Status)
	}

	if updatedDebts[2].Status != DebtStatusPaid || updatedDebts[2].RemainingBalance != 0 {
		t.Errorf("expected P3 BasicFit to be PAID, got balance %f status %s", updatedDebts[2].RemainingBalance, updatedDebts[2].Status)
	}

	// 2. Перевірка часткового погашення P2 (Monobank)
	if updatedDebts[1].RemainingBalance >= 20918.41 {
		t.Errorf("expected P2 Monobank balance to decrease, got %f", updatedDebts[1].RemainingBalance)
	}

	// 3. Перевірка позитивного чистого залишку вільного кешу
	if projection.FreeCashflowEUR <= 0 {
		t.Errorf("expected positive free cashflow, got %f", projection.FreeCashflowEUR)
	}
}

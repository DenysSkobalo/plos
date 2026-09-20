package finance

import (
	"math"
)

type CashflowEngine struct{}

func NewCashflowEngine() *CashflowEngine {
	return &CashflowEngine{}
}

func (e *CashflowEngine) ConvertToEURCents(amountCents int64, currency string, buyRate float64) int64 {
	if currency == "EUR" {
		return amountCents
	}
	if currency == "UAH" && buyRate > 0 {
		return int64(math.Round(float64(amountCents) / buyRate))
	}
	return amountCents
}

func (e *CashflowEngine) SimulateMonth(
	monthName string,
	incomes []IncomeItem,
	expenses []ExpenseItem,
	debts []DebtPayoffState,
	buyRate float64,
) (MonthlyProjection, []DebtPayoffState) {

	projection := MonthlyProjection{
		MonthName:        monthName,
		EffectiveBuyRate: buyRate,
	}

	for _, inc := range incomes {
		projection.TotalIncomeEURCents += e.ConvertToEURCents(inc.AmountCents, inc.Currency, buyRate)
	}

	for _, exp := range expenses {
		projection.TotalExpensesEURCents += e.ConvertToEURCents(exp.AmountCents, exp.Currency, buyRate)
	}

	availableCashEURCents := projection.TotalIncomeEURCents - projection.TotalExpensesEURCents

	currentDebts := make([]DebtPayoffState, len(debts))
	copy(currentDebts, debts)

	var totalDebtPaidEURCents int64

	// Phase A: Mandatory Minimum Payments
	for i := range currentDebts {
		if currentDebts[i].Status == DebtStatusPaid || currentDebts[i].RemainingBalanceCents <= 0 {
			continue
		}

		minPaymentEURCents := e.ConvertToEURCents(currentDebts[i].MinMonthlyPaymentUAHCents, "UAH", buyRate)
		if currentDebts[i].OriginalCurrency == "EUR" {
			minPaymentEURCents = currentDebts[i].MinMonthlyPaymentUAHCents
		}

		remainingDebtEURCents := e.ConvertToEURCents(currentDebts[i].RemainingBalanceCents, currentDebts[i].OriginalCurrency, buyRate)
		actualPaymentEURCents := minPaymentEURCents
		if remainingDebtEURCents < actualPaymentEURCents {
			actualPaymentEURCents = remainingDebtEURCents
		}
		if availableCashEURCents < actualPaymentEURCents {
			actualPaymentEURCents = availableCashEURCents
		}

		if actualPaymentEURCents > 0 {
			var paymentInOriginalCurrencyCents int64
			switch {
			case actualPaymentEURCents >= remainingDebtEURCents:
				paymentInOriginalCurrencyCents = currentDebts[i].RemainingBalanceCents
			case currentDebts[i].OriginalCurrency == "UAH":
				paymentInOriginalCurrencyCents = int64(math.Round(float64(actualPaymentEURCents) * buyRate))
			default:
				paymentInOriginalCurrencyCents = actualPaymentEURCents
			}

			currentDebts[i].RemainingBalanceCents -= paymentInOriginalCurrencyCents
			availableCashEURCents -= actualPaymentEURCents
			totalDebtPaidEURCents += actualPaymentEURCents

			if currentDebts[i].RemainingBalanceCents <= 0 {
				currentDebts[i].RemainingBalanceCents = 0
				currentDebts[i].Status = DebtStatusPaid
			}
		}
	}

	// Phase B: Cascade Surplus Distribution
	for i := range currentDebts {
		if availableCashEURCents <= 0 {
			break
		}
		if currentDebts[i].Status == DebtStatusPaid || currentDebts[i].RemainingBalanceCents <= 0 {
			continue
		}

		remainingDebtEURCents := e.ConvertToEURCents(currentDebts[i].RemainingBalanceCents, currentDebts[i].OriginalCurrency, buyRate)
		extraPaymentEURCents := availableCashEURCents
		if remainingDebtEURCents < extraPaymentEURCents {
			extraPaymentEURCents = remainingDebtEURCents
		}

		var extraPaymentInOriginalCurrencyCents int64
		switch {
		case extraPaymentEURCents >= remainingDebtEURCents:
			extraPaymentInOriginalCurrencyCents = currentDebts[i].RemainingBalanceCents
		case currentDebts[i].OriginalCurrency == "UAH":
			extraPaymentInOriginalCurrencyCents = int64(math.Round(float64(extraPaymentEURCents) * buyRate))
		default:
			extraPaymentInOriginalCurrencyCents = extraPaymentEURCents
		}

		currentDebts[i].RemainingBalanceCents -= extraPaymentInOriginalCurrencyCents
		availableCashEURCents -= extraPaymentEURCents
		totalDebtPaidEURCents += extraPaymentEURCents

		if currentDebts[i].RemainingBalanceCents <= 0 {
			currentDebts[i].RemainingBalanceCents = 0
			currentDebts[i].Status = DebtStatusPaid
		}
	}

	projection.DebtPaymentsEURCents = totalDebtPaidEURCents
	projection.FreeCashflowEURCents = availableCashEURCents
	projection.EndMonthDebtStates = currentDebts

	return projection, currentDebts
}

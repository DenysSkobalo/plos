package finance

import (
	"math"
)

type CashflowEngine struct{}

func NewCashflowEngine() *CashflowEngine {
	return &CashflowEngine{}
}

// ConvertToEUR конвертує суму в EUR за точним buyRate для UAH.
func (e *CashflowEngine) ConvertToEUR(amount float64, currency string, buyRate float64) float64 {
	if currency == "EUR" {
		return amount
	}
	if currency == "UAH" && buyRate > 0 {
		return amount / buyRate
	}
	return amount
}

// SimulateMonth виконує один крок помісячного каскадного погашення боргів.
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

	// 1. Обчислення загального доходу в EUR
	for _, inc := range incomes {
		projection.TotalIncomeEUR += e.ConvertToEUR(inc.Amount, inc.Currency, buyRate)
	}

	// 2. Обчислення фіксованих витрат в EUR
	for _, exp := range expenses {
		projection.TotalExpensesEUR += e.ConvertToEUR(exp.Amount, exp.Currency, buyRate)
	}

	availableCashEUR := projection.TotalIncomeEUR - projection.TotalExpensesEUR

	// Копіюємо стан боргів для симуляції
	currentDebts := make([]DebtPayoffState, len(debts))
	copy(currentDebts, debts)

	var totalDebtPaidEUR float64

	// 3. Фаза A: Внесення мінімальних обов'язкових платежів
	for i := range currentDebts {
		if currentDebts[i].Status == DebtStatusPaid || currentDebts[i].RemainingBalance <= 0 {
			continue
		}

		minPaymentEUR := e.ConvertToEUR(currentDebts[i].MinMonthlyPaymentUAH, "UAH", buyRate)
		if currentDebts[i].OriginalCurrency == "EUR" {
			minPaymentEUR = currentDebts[i].MinMonthlyPaymentUAH // якщо мінімальний платіж вказано безпосередньо в EUR
		}

		// Платіж не може перевищувати залишок боргу
		actualPaymentEUR := math.Min(minPaymentEUR, e.ConvertToEUR(currentDebts[i].RemainingBalance, currentDebts[i].OriginalCurrency, buyRate))
		actualPaymentEUR = math.Min(actualPaymentEUR, availableCashEUR)

		if actualPaymentEUR > 0 {
			paymentInOriginalCurrency := actualPaymentEUR
			if currentDebts[i].OriginalCurrency == "UAH" {
				paymentInOriginalCurrency = actualPaymentEUR * buyRate
			}

			currentDebts[i].RemainingBalance -= paymentInOriginalCurrency
			availableCashEUR -= actualPaymentEUR
			totalDebtPaidEUR += actualPaymentEUR

			if currentDebts[i].RemainingBalance <= 0.01 {
				currentDebts[i].RemainingBalance = 0
				currentDebts[i].Status = DebtStatusPaid
			}
		}
	}

	// 4. Фаза B: Каскадне спрямування залишку вільного кешфлоу на найвищий пріоритет (P1 -> P4)
	for i := range currentDebts {
		if availableCashEUR <= 0 {
			break
		}
		if currentDebts[i].Status == DebtStatusPaid || currentDebts[i].RemainingBalance <= 0 {
			continue
		}

		remainingDebtEUR := e.ConvertToEUR(currentDebts[i].RemainingBalance, currentDebts[i].OriginalCurrency, buyRate)
		extraPaymentEUR := math.Min(availableCashEUR, remainingDebtEUR)

		extraPaymentInOriginalCurrency := extraPaymentEUR
		if currentDebts[i].OriginalCurrency == "UAH" {
			extraPaymentInOriginalCurrency = extraPaymentEUR * buyRate
		}

		currentDebts[i].RemainingBalance -= extraPaymentInOriginalCurrency
		availableCashEUR -= extraPaymentEUR
		totalDebtPaidEUR += extraPaymentEUR

		if currentDebts[i].RemainingBalance <= 0.01 {
			currentDebts[i].RemainingBalance = 0
			currentDebts[i].Status = DebtStatusPaid
		}
	}

	projection.DebtPaymentsEUR = totalDebtPaidEUR
	projection.FreeCashflowEUR = availableCashEUR
	projection.EndMonthDebtStates = currentDebts

	return projection, currentDebts
}

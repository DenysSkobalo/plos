package finance

type IncomeItem struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ExpenseItem struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type DebtPayoffState struct {
	AccountID            string     `json:"account_id"`
	CreditorName         string     `json:"creditor_name"`
	Priority             int        `json:"priority"`
	OriginalCurrency     string     `json:"original_currency"`
	RemainingBalance     float64    `json:"remaining_balance"`
	MinMonthlyPaymentUAH float64    `json:"min_monthly_payment_uah"`
	Status               DebtStatus `json:"status"`
}

type MonthlyProjection struct {
	MonthName          string            `json:"month_name"` // Example: "2026-09"
	TotalIncomeEUR     float64           `json:"total_income_eur"`
	TotalExpensesEUR   float64           `json:"total_expenses_eur"`
	DebtPaymentsEUR    float64           `json:"debt_payments_eur"`
	FreeCashflowEUR    float64           `json:"free_cashflow_eur"`
	EffectiveBuyRate   float64           `json:"effective_buy_rate"`
	EndMonthDebtStates []DebtPayoffState `json:"end_month_debt_states"`
}

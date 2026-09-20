package finance

type IncomeItem struct {
	Name        string `json:"name"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type ExpenseItem struct {
	Name        string `json:"name"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type DebtPayoffState struct {
	AccountID                 string     `json:"account_id"`
	CreditorName              string     `json:"creditor_name"`
	Priority                  int        `json:"priority"`
	OriginalCurrency          string     `json:"original_currency"`
	RemainingBalanceCents     int64      `json:"remaining_balance_cents"`
	MinMonthlyPaymentUAHCents int64      `json:"min_monthly_payment_uah_cents"`
	Status                    DebtStatus `json:"status"`
}

type MonthlyProjection struct {
	MonthName             string            `json:"month_name"`
	TotalIncomeEURCents   int64             `json:"total_income_eur_cents"`
	TotalExpensesEURCents int64             `json:"total_expenses_eur_cents"`
	DebtPaymentsEURCents  int64             `json:"debt_payments_eur_cents"`
	FreeCashflowEURCents  int64             `json:"free_cashflow_eur_cents"`
	EffectiveBuyRate      float64           `json:"effective_buy_rate"`
	EndMonthDebtStates    []DebtPayoffState `json:"end_month_debt_states"`
}

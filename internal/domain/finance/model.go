package finance

import "time"

type AccountType string

const (
	AccountTypeIncomeSource    AccountType = "INCOME_SOURCE"
	AccountTypeExpenseCategory AccountType = "EXPENSE_CATEGORY"
	AccountTypeDebt            AccountType = "DEBT"
	AccountTypeAsset           AccountType = "ASSET"
)

type DebtStatus string

const (
	DebtStatusOverdue DebtStatus = "OVERDUE"
	DebtStatusActive  DebtStatus = "ACTIVE"
	DebtStatusPaid    DebtStatus = "PAID"
)

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

type ExchangeRate struct {
	ID             int64     `json:"id"`
	Source         string    `json:"source"` // 'MONO', 'PUMB', 'MANUAL'
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	BuyRate        float64   `json:"buy_rate"`
	SellRate       float64   `json:"sell_rate"`
	IsManual       bool      `json:"is_manual"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Account struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           AccountType `json:"type"`
	Currency       string      `json:"currency"`
	InitialBalance float64     `json:"initial_balance"`
	CurrentBalance float64     `json:"current_balance"`
	CreatedAt      time.Time   `json:"created_at"`
}

type DebtMetadata struct {
	AccountID            string     `json:"account_id"`
	Priority             int        `json:"priority"`
	CreditorName         string     `json:"creditor_name"`
	OriginalCurrency     string     `json:"original_currency"`
	OriginalAmount       float64    `json:"original_amount"`
	MinMonthlyPaymentUAH float64    `json:"min_monthly_payment_uah"`
	DeadlineDate         *string    `json:"deadline_date,omitempty"` // YYYY-MM-DD
	Status               DebtStatus `json:"status"`
}

type Transaction struct {
	ID                   string    `json:"id"`
	SourceAccountID      *string   `json:"source_account_id,omitempty"`
	DestinationAccountID *string   `json:"destination_account_id,omitempty"`
	Amount               float64   `json:"amount"`
	Currency             string    `json:"currency"`
	ExchangeRateApplied  float64   `json:"exchange_rate_applied"`
	Description          string    `json:"description"`
	ExecutionDate        string    `json:"execution_date"` // YYYY-MM-DD
	CreatedAt            time.Time `json:"created_at"`
}

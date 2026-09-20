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
	Source         string    `json:"source"`
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	BuyRate        float64   `json:"buy_rate"`
	SellRate       float64   `json:"sell_rate"`
	IsManual       bool      `json:"is_manual"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Account struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	Type                AccountType `json:"type"`
	Currency            string      `json:"currency"`
	InitialBalanceCents int64       `json:"initial_balance_cents"`
	CurrentBalanceCents int64       `json:"current_balance_cents"`
	CreatedAt           time.Time   `json:"created_at"`
}

type DebtMetadata struct {
	AccountID                 string     `json:"account_id"`
	Priority                  int        `json:"priority"`
	CreditorName              string     `json:"creditor_name"`
	OriginalCurrency          string     `json:"original_currency"`
	OriginalAmountCents       int64      `json:"original_amount_cents"`
	MinMonthlyPaymentUAHCents int64      `json:"min_monthly_payment_uah_cents"`
	DeadlineDate              *string    `json:"deadline_date,omitempty"`
	Status                    DebtStatus `json:"status"`
}

type Transaction struct {
	ID                   string    `json:"id"`
	SourceAccountID      *string   `json:"source_account_id,omitempty"`
	DestinationAccountID *string   `json:"destination_account_id,omitempty"`
	AmountCents          int64     `json:"amount_cents"`
	Currency             string    `json:"currency"`
	ExchangeRateApplied  float64   `json:"exchange_rate_applied"`
	Description          string    `json:"description"`
	ExecutionDate        string    `json:"execution_date"`
	CreatedAt            time.Time `json:"created_at"`
}

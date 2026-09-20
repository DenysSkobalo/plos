package finance

import "context"

type Repository interface {
	// Currencies & Rates
	AddCurrency(ctx context.Context, currency Currency) error
	GetCurrencies(ctx context.Context) ([]Currency, error)
	SaveExchangeRate(ctx context.Context, rate *ExchangeRate) error
	GetLatestRates(ctx context.Context) ([]ExchangeRate, error)

	// Accounts & Debts
	CreateAccount(ctx context.Context, acc *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context) ([]Account, error)
	SetDebtMetadata(ctx context.Context, meta *DebtMetadata) error
	GetDebtMetadata(ctx context.Context, accountID string) (*DebtMetadata, error)
	GetActiveDebtsOrderedByPriority(ctx context.Context) ([]DebtMetadata, error)

	// Transactions
	CreateTransaction(ctx context.Context, tx *Transaction) error
	ProcessTransaction(ctx context.Context, tx *Transaction) error
	GetTransactions(ctx context.Context, limit int) ([]Transaction, error)
}

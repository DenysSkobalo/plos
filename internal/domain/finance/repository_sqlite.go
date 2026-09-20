package finance

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) AddCurrency(ctx context.Context, currency Currency) error {
	query := `INSERT INTO currencies (code, symbol) VALUES (?, ?) ON CONFLICT(code) DO UPDATE SET symbol=excluded.symbol`
	_, err := r.db.ExecContext(ctx, query, currency.Code, currency.Symbol)
	if err != nil {
		return fmt.Errorf("failed to add currency: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetCurrencies(ctx context.Context) ([]Currency, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code, symbol FROM currencies ORDER BY code ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query currencies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []Currency
	for rows.Next() {
		var c Currency
		if err := rows.Scan(&c.Code, &c.Symbol); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *SQLiteRepository) SaveExchangeRate(ctx context.Context, rate *ExchangeRate) error {
	query := `
		INSERT INTO exchange_rates (source, base_currency, target_currency, buy_rate, sell_rate, is_manual, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := r.db.ExecContext(ctx, query, rate.Source, rate.BaseCurrency, rate.TargetCurrency, rate.BuyRate, rate.SellRate, rate.IsManual)
	if err != nil {
		return fmt.Errorf("failed to save exchange rate: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetLatestRates(ctx context.Context) ([]ExchangeRate, error) {
	query := `
		SELECT id, source, base_currency, target_currency, buy_rate, sell_rate, is_manual, updated_at
		FROM exchange_rates
		WHERE id IN (
			SELECT MAX(id) FROM exchange_rates GROUP BY source, base_currency, target_currency
		)
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest rates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var rates []ExchangeRate
	for rows.Next() {
		var rate ExchangeRate
		if err := rows.Scan(&rate.ID, &rate.Source, &rate.BaseCurrency, &rate.TargetCurrency, &rate.BuyRate, &rate.SellRate, &rate.IsManual, &rate.UpdatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, rows.Err()
}

func (r *SQLiteRepository) CreateAccount(ctx context.Context, acc *Account) error {
	query := `
		INSERT INTO accounts (id, name, type, currency, initial_balance, current_balance)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, acc.ID, acc.Name, string(acc.Type), acc.Currency, acc.InitialBalance, acc.CurrentBalance)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	query := `SELECT id, name, type, currency, initial_balance, current_balance, created_at FROM accounts WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var acc Account
	var accType string
	if err := row.Scan(&acc.ID, &acc.Name, &accType, &acc.Currency, &acc.InitialBalance, &acc.CurrentBalance, &acc.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan account: %w", err)
	}
	acc.Type = AccountType(accType)
	return &acc, nil
}

func (r *SQLiteRepository) GetAccounts(ctx context.Context) ([]Account, error) {
	query := `SELECT id, name, type, currency, initial_balance, current_balance, created_at FROM accounts ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var accounts []Account
	for rows.Next() {
		var acc Account
		var accType string
		if err := rows.Scan(&acc.ID, &acc.Name, &accType, &acc.Currency, &acc.InitialBalance, &acc.CurrentBalance, &acc.CreatedAt); err != nil {
			return nil, err
		}
		acc.Type = AccountType(accType)
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

func (r *SQLiteRepository) SetDebtMetadata(ctx context.Context, meta *DebtMetadata) error {
	query := `
		INSERT INTO debt_metadata (account_id, priority, creditor_name, original_currency, original_amount, min_monthly_payment_uah, deadline_date, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id) DO UPDATE SET
			priority=excluded.priority,
			creditor_name=excluded.creditor_name,
			original_currency=excluded.original_currency,
			original_amount=excluded.original_amount,
			min_monthly_payment_uah=excluded.min_monthly_payment_uah,
			deadline_date=excluded.deadline_date,
			status=excluded.status
	`
	_, err := r.db.ExecContext(ctx, query,
		meta.AccountID, meta.Priority, meta.CreditorName, meta.OriginalCurrency,
		meta.OriginalAmount, meta.MinMonthlyPaymentUAH, meta.DeadlineDate, string(meta.Status),
	)
	if err != nil {
		return fmt.Errorf("failed to set debt metadata: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetDebtMetadata(ctx context.Context, accountID string) (*DebtMetadata, error) {
	query := `SELECT account_id, priority, creditor_name, original_currency, original_amount, min_monthly_payment_uah, deadline_date, status FROM debt_metadata WHERE account_id = ?`
	row := r.db.QueryRowContext(ctx, query, accountID)

	var meta DebtMetadata
	var status string
	if err := row.Scan(&meta.AccountID, &meta.Priority, &meta.CreditorName, &meta.OriginalCurrency, &meta.OriginalAmount, &meta.MinMonthlyPaymentUAH, &meta.DeadlineDate, &status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan debt metadata: %w", err)
	}
	meta.Status = DebtStatus(status)
	return &meta, nil
}

func (r *SQLiteRepository) GetActiveDebtsOrderedByPriority(ctx context.Context) ([]DebtMetadata, error) {
	query := `
		SELECT account_id, priority, creditor_name, original_currency, original_amount, min_monthly_payment_uah, deadline_date, status
		FROM debt_metadata
		WHERE status != 'PAID'
		ORDER BY priority ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active debts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []DebtMetadata
	for rows.Next() {
		var meta DebtMetadata
		var status string
		if err := rows.Scan(&meta.AccountID, &meta.Priority, &meta.CreditorName, &meta.OriginalCurrency, &meta.OriginalAmount, &meta.MinMonthlyPaymentUAH, &meta.DeadlineDate, &status); err != nil {
			return nil, err
		}
		meta.Status = DebtStatus(status)
		list = append(list, meta)
	}
	return list, rows.Err()
}

func (r *SQLiteRepository) CreateTransaction(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO transactions (id, source_account_id, destination_account_id, amount, currency, exchange_rate_applied, description, execution_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, tx.ID, tx.SourceAccountID, tx.DestinationAccountID, tx.Amount, tx.Currency, tx.ExchangeRateApplied, tx.Description, tx.ExecutionDate)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetTransactions(ctx context.Context, limit int) ([]Transaction, error) {
	query := `
		SELECT id, source_account_id, destination_account_id, amount, currency, exchange_rate_applied, description, execution_date, created_at
		FROM transactions
		ORDER BY execution_date DESC, created_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.SourceAccountID, &tx.DestinationAccountID, &tx.Amount, &tx.Currency, &tx.ExchangeRateApplied, &tx.Description, &tx.ExecutionDate, &tx.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, tx)
	}
	return list, rows.Err()
}

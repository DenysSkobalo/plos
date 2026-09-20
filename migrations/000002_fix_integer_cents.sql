CREATE TABLE accounts_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    currency TEXT NOT NULL,
    initial_balance_cents INTEGER NOT NULL DEFAULT 0,
    current_balance_cents INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (currency) REFERENCES currencies(code)
);

INSERT INTO accounts_new (id, name, type, currency, initial_balance_cents, current_balance_cents, created_at)
SELECT id, name, type, currency, CAST(ROUND(initial_balance * 100) AS INTEGER), CAST(ROUND(current_balance * 100) AS INTEGER), created_at
FROM accounts;

DROP TABLE accounts;
ALTER TABLE accounts_new RENAME TO accounts;

CREATE TABLE debt_metadata_new (
    account_id TEXT PRIMARY KEY,
    priority INTEGER NOT NULL,
    creditor_name TEXT NOT NULL,
    original_currency TEXT NOT NULL,
    original_amount_cents INTEGER NOT NULL,
    min_monthly_payment_uah_cents INTEGER NOT NULL,
    deadline_date TEXT,
    status TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (original_currency) REFERENCES currencies(code)
);

INSERT INTO debt_metadata_new (account_id, priority, creditor_name, original_currency, original_amount_cents, min_monthly_payment_uah_cents, deadline_date, status)
SELECT account_id, priority, creditor_name, original_currency, CAST(ROUND(original_amount * 100) AS INTEGER), CAST(ROUND(min_monthly_payment_uah * 100) AS INTEGER), deadline_date, status
FROM debt_metadata;

DROP TABLE debt_metadata;
ALTER TABLE debt_metadata_new RENAME TO debt_metadata;

CREATE TABLE transactions_new (
    id TEXT PRIMARY KEY,
    source_account_id TEXT,
    destination_account_id TEXT,
    amount_cents INTEGER NOT NULL,
    currency TEXT NOT NULL,
    exchange_rate_applied REAL DEFAULT 1.0,
    description TEXT,
    execution_date TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_account_id) REFERENCES accounts(id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(id),
    FOREIGN KEY (currency) REFERENCES currencies(code)
);

INSERT INTO transactions_new (id, source_account_id, destination_account_id, amount_cents, currency, exchange_rate_applied, description, execution_date, created_at)
SELECT id, source_account_id, destination_account_id, CAST(ROUND(amount * 100) AS INTEGER), currency, exchange_rate_applied, description, execution_date, created_at
FROM transactions;

DROP TABLE transactions;
ALTER TABLE transactions_new RENAME TO transactions;

CREATE INDEX IF NOT EXISTS idx_transactions_execution_date ON transactions(execution_date DESC);

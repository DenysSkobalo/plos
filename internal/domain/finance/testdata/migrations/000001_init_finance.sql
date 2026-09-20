-- Schema Versioning Tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Currencies Reference Table
CREATE TABLE IF NOT EXISTS currencies (
    code TEXT PRIMARY KEY,
    symbol TEXT NOT NULL
);

-- Base ISO Currencies Seed
INSERT OR IGNORE INTO currencies (code, symbol) VALUES 
    ('EUR', '€'),
    ('UAH', '₴'),
    ('USD', '$');

-- Exchange Rates
CREATE TABLE IF NOT EXISTS exchange_rates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL, -- 'MONO', 'PUMB', 'MANUAL'
    base_currency TEXT NOT NULL,
    target_currency TEXT NOT NULL,
    buy_rate REAL NOT NULL,
    sell_rate REAL NOT NULL,
    is_manual BOOLEAN NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (base_currency) REFERENCES currencies(code),
    FOREIGN KEY (target_currency) REFERENCES currencies(code)
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_lookup 
ON exchange_rates(source, base_currency, target_currency, is_manual, updated_at DESC);

-- Accounts & Debt Categories
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'INCOME_SOURCE', 'EXPENSE_CATEGORY', 'DEBT', 'ASSET'
    currency TEXT NOT NULL,
    initial_balance REAL DEFAULT 0.0,
    current_balance REAL DEFAULT 0.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (currency) REFERENCES currencies(code)
);

-- Debt Metadata (Cascade Payoff Attributes)
CREATE TABLE IF NOT EXISTS debt_metadata (
    account_id TEXT PRIMARY KEY,
    priority INTEGER NOT NULL,
    creditor_name TEXT NOT NULL,
    original_currency TEXT NOT NULL,
    original_amount REAL NOT NULL,
    min_monthly_payment_uah REAL NOT NULL,
    deadline_date TEXT,
    status TEXT NOT NULL, -- 'OVERDUE', 'ACTIVE', 'PAID'
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (original_currency) REFERENCES currencies(code)
);

-- Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    source_account_id TEXT,
    destination_account_id TEXT,
    amount REAL NOT NULL,
    currency TEXT NOT NULL,
    exchange_rate_applied REAL DEFAULT 1.0,
    description TEXT,
    execution_date TEXT NOT NULL, -- YYYY-MM-DD
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_account_id) REFERENCES accounts(id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(id),
    FOREIGN KEY (currency) REFERENCES currencies(code)
);

CREATE INDEX IF NOT EXISTS idx_transactions_execution_date 
ON transactions(execution_date DESC);

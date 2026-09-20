#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"

echo "=== 1. POST /currencies ==="
curl -s -X POST "${BASE_URL}/currencies" \
  -H "Content-Type: application/json" \
  -d '{"code":"PLN","symbol":"zł"}' -w "\nHTTP Status: %{http_code}\n\n"

echo "=== 2. GET /currencies ==="
curl -s -X GET "${BASE_URL}/currencies" | jq .

echo "=== 3. POST /rates/manual ==="
curl -s -X POST "${BASE_URL}/rates/manual" \
  -H "Content-Type: application/json" \
  -d '{
    "base_currency": "EUR",
    "target_currency": "UAH",
    "buy_rate": 51.0700,
    "sell_rate": 51.7706
  }' -w "\nHTTP Status: %{http_code}\n\n"

echo "=== 4. POST /rates/sync (Monobank API) ==="
curl -s -X POST "${BASE_URL}/rates/sync" -w "HTTP Status: %{http_code}\n\n"

echo "=== 5. GET /rates ==="
curl -s -X GET "${BASE_URL}/rates" | jq .

echo "=== 6. POST /accounts (Debt Account) ==="
curl -s -X POST "${BASE_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "acc-mfo-1",
    "name": "Moneyveo",
    "type": "DEBT",
    "currency": "UAH",
    "initial_balance_cents": -772192,
    "current_balance_cents": -772192
  }' -w "\nHTTP Status: %{http_code}\n\n"

echo "=== 7. GET /accounts ==="
curl -s -X GET "${BASE_URL}/accounts" | jq .

echo "=== 8. POST /debts ==="
curl -s -X POST "${BASE_URL}/debts" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "acc-mfo-1",
    "priority": 1,
    "creditor_name": "Moneyveo",
    "original_currency": "UAH",
    "original_amount_cents": 772192,
    "min_monthly_payment_uah_cents": 772192,
    "status": "OVERDUE"
  }' -w "\nHTTP Status: %{http_code}\n\n"

echo "=== 9. GET /debts ==="
curl -s -X GET "${BASE_URL}/debts" | jq .

echo "=== 10. POST /cashflow/simulate ==="
curl -s -X POST "${BASE_URL}/cashflow/simulate" \
  -H "Content-Type: application/json" \
  -d '{
    "month_name": "2026-09",
    "buy_rate": 51.0700,
    "incomes": [
      {"name": "Salary September", "amount_cents": 123000, "currency": "EUR"}
    ],
    "expenses": [
      {"name": "Mobile", "amount_cents": 1220, "currency": "EUR"},
      {"name": "TCL Pass", "amount_cents": 2050, "currency": "EUR"}
    ]
  }' | jq .

echo "=== 11. GET /export/debts.csv ==="
curl -s -i -X GET "${BASE_URL}/export/debts.csv"

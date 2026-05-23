-- name: GetAccountByCode :one
SELECT * FROM accounts WHERE code = $1 AND is_active = TRUE;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE is_active = TRUE ORDER BY code;

-- name: CreateJournalEntry :one
INSERT INTO journal_entries (posted_on, memo, source, source_ref, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: AddJournalLine :one
INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, memo)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListJournalEntries :many
SELECT * FROM journal_entries
ORDER BY posted_on DESC, created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetJournalEntry :one
SELECT * FROM journal_entries WHERE id = $1;

-- name: GetJournalLines :many
SELECT jel.*, a.code AS account_code, a.name AS account_name
FROM journal_entry_lines jel
JOIN accounts a ON a.id = jel.account_id
WHERE jel.journal_entry_id = $1
ORDER BY jel.debit DESC;

-- ─── Expense categories ───────────────────────────────────────────────────────

-- name: ListExpenseCategories :many
SELECT * FROM expense_categories WHERE is_active = TRUE ORDER BY name;

-- name: CreateExpenseCategory :one
INSERT INTO expense_categories (name, account_id)
VALUES ($1, $2)
RETURNING *;

-- ─── Expenses ─────────────────────────────────────────────────────────────────

-- name: CreateExpense :one
INSERT INTO expenses (expense_category_id, location_id, amount, paid_on, note, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateExpenseJournalRef :exec
UPDATE expenses SET journal_entry_id = $2 WHERE id = $1;

-- name: ListExpenses :many
SELECT e.*, ec.name AS category_name
FROM expenses e
JOIN expense_categories ec ON ec.id = e.expense_category_id
WHERE e.location_id = $1
  AND ($2::date IS NULL OR e.paid_on >= $2)
  AND ($3::date IS NULL OR e.paid_on <= $3)
ORDER BY e.paid_on DESC
LIMIT 50;

-- ─── Tax records ──────────────────────────────────────────────────────────────

-- name: UpsertTaxRecord :one
INSERT INTO tax_records (location_id, kind, base, amount, period_start, period_end)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: ListTaxRecords :many
SELECT * FROM tax_records
WHERE location_id = $1
ORDER BY period_start DESC
LIMIT 20;

-- name: MarkTaxPaid :exec
UPDATE tax_records SET is_paid = TRUE, updated_at = NOW() WHERE id = $1;

DROP TRIGGER IF EXISTS trg_transaction_entries_immutable ON transaction_entries;
DROP TRIGGER IF EXISTS trg_transactions_immutable ON transactions;
DROP FUNCTION IF EXISTS prevent_ledger_mutation();
DROP INDEX IF EXISTS ux_transactions_corrects_once;
DROP INDEX IF EXISTS ux_transactions_reverses_once;

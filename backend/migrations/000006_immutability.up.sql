CREATE UNIQUE INDEX ux_transactions_reverses_once
    ON transactions (reverses_transaction_id)
    WHERE reverses_transaction_id IS NOT NULL;

CREATE UNIQUE INDEX ux_transactions_corrects_once
    ON transactions (corrects_transaction_id)
    WHERE corrects_transaction_id IS NOT NULL;

CREATE OR REPLACE FUNCTION prevent_ledger_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'ledger records are immutable';
END;
$$;

CREATE TRIGGER trg_transactions_immutable
    BEFORE UPDATE OR DELETE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION prevent_ledger_mutation();

CREATE TRIGGER trg_transaction_entries_immutable
    BEFORE UPDATE OR DELETE ON transaction_entries
    FOR EACH ROW
    EXECUTE FUNCTION prevent_ledger_mutation();

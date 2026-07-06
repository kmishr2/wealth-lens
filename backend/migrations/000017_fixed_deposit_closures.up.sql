ALTER TABLE fixed_deposits
    ADD CONSTRAINT uq_fixed_deposits_scope_currency UNIQUE (id, portfolio_id, account_id, currency);

CREATE TABLE fixed_deposit_closures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    fixed_deposit_id uuid NOT NULL UNIQUE,
    portfolio_id uuid NOT NULL,
    account_id uuid NOT NULL,
    closing_transaction_id uuid NOT NULL UNIQUE,
    closure_type text NOT NULL CHECK (closure_type IN ('maturity', 'premature')),
    closed_at date NOT NULL,
    proceeds numeric(28,4) NOT NULL CHECK (proceeds > 0),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    note text NOT NULL DEFAULT '',
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_fd_closures_fixed_deposit_scope FOREIGN KEY (fixed_deposit_id, portfolio_id, account_id, currency) REFERENCES fixed_deposits (id, portfolio_id, account_id, currency) ON DELETE RESTRICT,
    CONSTRAINT fk_fd_closures_account_portfolio FOREIGN KEY (account_id, portfolio_id) REFERENCES accounts (id, portfolio_id) ON DELETE RESTRICT,
    CONSTRAINT fk_fd_closures_transaction FOREIGN KEY (closing_transaction_id, portfolio_id, account_id) REFERENCES transactions (id, portfolio_id, account_id) ON DELETE RESTRICT,
    CONSTRAINT fk_fd_closures_creator_portfolio FOREIGN KEY (portfolio_id, created_by_user_id) REFERENCES portfolios (id, user_id) ON DELETE RESTRICT
);
CREATE INDEX idx_fd_closures_portfolio_account ON fixed_deposit_closures (portfolio_id, account_id, closed_at DESC);
CREATE TRIGGER trg_fixed_deposit_closures_immutable BEFORE UPDATE OR DELETE ON fixed_deposit_closures FOR EACH ROW EXECUTE FUNCTION prevent_ledger_mutation();

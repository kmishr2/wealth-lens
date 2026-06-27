CREATE TABLE transactions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE RESTRICT,
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    transaction_type text NOT NULL CHECK (transaction_type IN ('deposit', 'withdrawal', 'buy', 'sell', 'fee', 'tax', 'transfer', 'reversal')),
    occurred_at timestamptz NOT NULL,
    description text NOT NULL DEFAULT '',
    idempotency_key text,
    reverses_transaction_id uuid REFERENCES transactions(id) ON DELETE RESTRICT,
    corrects_transaction_id uuid REFERENCES transactions(id) ON DELETE RESTRICT,
    correction_group_id uuid,
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_portfolio_id ON transactions (portfolio_id);
CREATE INDEX idx_transactions_account_id ON transactions (account_id);
CREATE INDEX idx_transactions_created_by_user_id ON transactions (created_by_user_id);
CREATE INDEX idx_transactions_occurred_order ON transactions (portfolio_id, occurred_at DESC, created_at DESC, id DESC);
CREATE INDEX idx_transactions_reverses_transaction_id ON transactions (reverses_transaction_id);
CREATE INDEX idx_transactions_corrects_transaction_id ON transactions (corrects_transaction_id);
CREATE INDEX idx_transactions_correction_group_id ON transactions (correction_group_id);
CREATE UNIQUE INDEX ux_transactions_portfolio_idempotency
    ON transactions (portfolio_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE transaction_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id uuid NOT NULL REFERENCES transactions(id) ON DELETE RESTRICT,
    entry_kind text NOT NULL CHECK (entry_kind IN ('cash', 'asset', 'fee', 'tax')),
    asset_id uuid REFERENCES assets(id) ON DELETE RESTRICT,
    quantity numeric(28,10),
    amount numeric(28,4),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (entry_kind <> 'asset' OR (asset_id IS NOT NULL AND quantity IS NOT NULL AND quantity <> 0)),
    CHECK (entry_kind = 'asset' OR (amount IS NOT NULL AND amount <> 0))
);

CREATE INDEX idx_transaction_entries_transaction_id ON transaction_entries (transaction_id);
CREATE INDEX idx_transaction_entries_asset_id ON transaction_entries (asset_id);

ALTER TABLE assets DROP CONSTRAINT assets_asset_class_check;
ALTER TABLE assets
    ADD CONSTRAINT assets_asset_class_check CHECK (
        asset_class IN ('cash', 'equity', 'fund', 'bond', 'fixed_deposit', 'crypto', 'real_estate', 'commodity', 'alternative', 'other')
    );

CREATE TABLE fixed_deposits (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL,
    account_id uuid NOT NULL,
    asset_id uuid NOT NULL UNIQUE REFERENCES assets(id) ON DELETE RESTRICT,
    opening_transaction_id uuid NOT NULL UNIQUE,
    name text NOT NULL CHECK (length(trim(name)) > 0),
    bank_reference text NOT NULL DEFAULT '',
    principal numeric(28,4) NOT NULL CHECK (principal > 0),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    annual_interest_rate numeric(9,6) NOT NULL CHECK (annual_interest_rate > 0 AND annual_interest_rate <= 100),
    start_date date NOT NULL,
    maturity_date date NOT NULL CHECK (maturity_date > start_date),
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_fixed_deposits_account_portfolio
        FOREIGN KEY (account_id, portfolio_id)
        REFERENCES accounts (id, portfolio_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_fixed_deposits_opening_transaction
        FOREIGN KEY (opening_transaction_id, portfolio_id, account_id)
        REFERENCES transactions (id, portfolio_id, account_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_fixed_deposits_creator_portfolio
        FOREIGN KEY (portfolio_id, created_by_user_id)
        REFERENCES portfolios (id, user_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_fixed_deposits_portfolio_account
    ON fixed_deposits (portfolio_id, account_id, start_date DESC, id DESC);

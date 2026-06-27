CREATE TABLE portfolios (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    base_currency char(3) NOT NULL CHECK (base_currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_portfolios_user_id ON portfolios (user_id);
CREATE INDEX idx_portfolios_deleted_at ON portfolios (deleted_at);
CREATE UNIQUE INDEX ux_portfolios_user_name_active
    ON portfolios (user_id, lower(name))
    WHERE deleted_at IS NULL;

CREATE TABLE accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    name text NOT NULL,
    account_type text NOT NULL CHECK (account_type IN ('brokerage', 'retirement', 'bank', 'wallet', 'other')),
    institution_name text NOT NULL DEFAULT '',
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_accounts_portfolio_id ON accounts (portfolio_id);
CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at);
CREATE UNIQUE INDEX ux_accounts_portfolio_name_active
    ON accounts (portfolio_id, lower(name))
    WHERE deleted_at IS NULL;

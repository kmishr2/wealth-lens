CREATE TABLE assets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol text NOT NULL CHECK (symbol = upper(symbol)),
    name text NOT NULL,
    asset_class text NOT NULL CHECK (asset_class IN ('cash', 'equity', 'fund', 'bond', 'crypto', 'real_estate', 'commodity', 'alternative', 'other')),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    exchange text NOT NULL DEFAULT '' CHECK (exchange = upper(exchange)),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_assets_symbol_exchange_currency
    ON assets (symbol, exchange, currency);
CREATE INDEX idx_assets_asset_class ON assets (asset_class);
CREATE INDEX idx_assets_is_active ON assets (is_active);

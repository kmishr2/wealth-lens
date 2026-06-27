CREATE TABLE asset_prices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    price numeric(28,10) NOT NULL CHECK (price > 0),
    currency char(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    priced_at timestamptz NOT NULL,
    source text NOT NULL DEFAULT 'manual' CHECK (length(trim(source)) > 0),
    note text NOT NULL DEFAULT '',
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_asset_prices_asset_priced_at
    ON asset_prices (asset_id, priced_at DESC, created_at DESC, id DESC);
CREATE INDEX idx_asset_prices_created_by_user_id
    ON asset_prices (created_by_user_id);

CREATE OR REPLACE FUNCTION prevent_asset_price_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'asset price snapshots are immutable';
END;
$$;

CREATE TRIGGER trg_asset_prices_immutable
    BEFORE UPDATE OR DELETE ON asset_prices
    FOR EACH ROW
    EXECUTE FUNCTION prevent_asset_price_mutation();

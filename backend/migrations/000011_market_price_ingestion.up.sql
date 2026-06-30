CREATE TABLE asset_identifiers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider = lower(provider) AND length(trim(provider)) > 0),
    identifier text NOT NULL CHECK (length(trim(identifier)) > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (asset_id, provider),
    UNIQUE (provider, identifier)
);

CREATE INDEX idx_asset_identifiers_asset_id ON asset_identifiers (asset_id);

ALTER TABLE asset_prices
    ALTER COLUMN created_by_user_id DROP NOT NULL,
    ADD COLUMN market_date date,
    ADD CONSTRAINT ck_asset_prices_origin
        CHECK (created_by_user_id IS NOT NULL OR source <> 'manual');

CREATE UNIQUE INDEX ux_asset_prices_automated_market_date
    ON asset_prices (asset_id, market_date, source)
    WHERE market_date IS NOT NULL;


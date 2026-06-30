DROP INDEX IF EXISTS ux_asset_prices_automated_market_date;

-- Version 8 cannot represent system-ingested prices because it requires a
-- user creator. Remove only those rows while the immutable trigger is
-- temporarily disabled, then restore the trigger for the version-8 schema.
DROP TRIGGER IF EXISTS trg_asset_prices_immutable ON asset_prices;
DELETE FROM asset_prices WHERE created_by_user_id IS NULL;

ALTER TABLE asset_prices
    DROP CONSTRAINT IF EXISTS ck_asset_prices_origin,
    DROP COLUMN IF EXISTS market_date,
    ALTER COLUMN created_by_user_id SET NOT NULL;

DROP TABLE IF EXISTS asset_identifiers;

CREATE TRIGGER trg_asset_prices_immutable
    BEFORE UPDATE OR DELETE ON asset_prices
    FOR EACH ROW
    EXECUTE FUNCTION prevent_asset_price_mutation();

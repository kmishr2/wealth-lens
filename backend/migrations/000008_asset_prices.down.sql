DROP TRIGGER IF EXISTS trg_asset_prices_immutable ON asset_prices;
DROP FUNCTION IF EXISTS prevent_asset_price_mutation();
DROP TABLE IF EXISTS asset_prices;

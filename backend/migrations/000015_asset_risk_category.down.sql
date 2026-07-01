DROP INDEX IF EXISTS idx_assets_risk_category;
ALTER TABLE assets DROP COLUMN IF EXISTS risk_category;

DROP TABLE IF EXISTS fixed_deposits;

ALTER TABLE assets DROP CONSTRAINT assets_asset_class_check;
ALTER TABLE assets
    ADD CONSTRAINT assets_asset_class_check CHECK (
        asset_class IN ('cash', 'equity', 'fund', 'bond', 'crypto', 'real_estate', 'commodity', 'alternative', 'other')
    );

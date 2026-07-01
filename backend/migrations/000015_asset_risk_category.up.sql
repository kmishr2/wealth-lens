ALTER TABLE assets
    ADD COLUMN risk_category text
    CHECK (risk_category IN ('equity', 'debt', 'cash_other'));

UPDATE assets SET risk_category = 'equity' WHERE asset_class = 'equity';
UPDATE assets SET risk_category = 'debt' WHERE asset_class = 'bond';
UPDATE assets SET risk_category = 'cash_other' WHERE asset_class = 'cash';

CREATE INDEX idx_assets_risk_category ON assets (risk_category);

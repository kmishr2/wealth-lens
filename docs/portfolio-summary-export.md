# Portfolio Summary CSV Export

`GET /portfolios/:portfolioId/exports/summary` downloads a current portfolio
summary from the authenticated frontend.

The CSV uses one fixed schema and identifies each row with `record_type`:

- `portfolio_total`: one current total per currency;
- `asset_holding`: ledger-derived quantities with available market value and allocation;
- `cash_balance`: ledger-derived cash per currency;
- `asset_class_allocation`: current asset-class allocation per currency;
- `missing_price`: explicit disclosure for unvalued holdings;
- `metric_definition`: formula, assumptions, required inputs, and explanation.

Currencies are never combined or converted. Missing prices remain visible and
are not replaced with inferred values. Spreadsheet formula prefixes in text
fields are neutralized, every field is CSV-quoted, and responses disable caching.

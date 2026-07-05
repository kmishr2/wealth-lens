# Transaction CSV Import

CSV imports are account-scoped and atomic. Every row is validated before any
ledger event is written. A failed row rejects the entire file.

Required header, in this exact order:

```csv
transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key
```

- `transaction_type`: `deposit`, `withdrawal`, `buy`, `sell`, `fee`, or `tax`
- `occurred_at`: RFC3339 timestamp, for example `2026-07-05T10:30:00+05:30`
- `asset_id` and `quantity`: required only for `buy` and `sell`
- `amount`: positive cash amount; the backend applies the ledger sign
- `currency`: three-letter currency code matching the account and asset rules
- `idempotency_key`: required and unique within the portfolio

Files are limited to 1,000 data rows and 2 MiB. Transfers, reversals, and
corrections must use their dedicated workflows and are not importable.

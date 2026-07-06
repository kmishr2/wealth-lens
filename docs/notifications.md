# Deterministic Notices

Notices are derived from source records when requested. They are not stored as
portfolio state and do not contain predictions or recommendations.

## Fixed-deposit maturity

`GET /api/v1/notifications` returns an open fixed-deposit notice when its
maturity date is no more than 30 calendar days after the as-of date. Closed
deposits are excluded. An unclosed deposit remains visible after maturity.

The optional `as_of_date=YYYY-MM-DD` query parameter makes the result
reproducible. When omitted, the current UTC date is used.

Statuses are explicit:

- `upcoming`: 8–30 calendar days remain;
- `urgent`: 1–7 calendar days remain;
- `due`: maturity is today;
- `overdue`: maturity has passed without a recorded closure.

Each response includes the trigger rule, explanation, source entity, event
date, as-of date, and portfolio/account links. The stable identifier is derived
from the notice kind and fixed-deposit ID.

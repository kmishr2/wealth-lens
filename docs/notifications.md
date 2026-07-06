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

## Goal target dates

An active goal appears when its target date is within the same 30-day window or
is overdue. Completed, archived, deleted, and goals marked reached by their
latest monthly snapshot on or before the as-of date are excluded. If no monthly
snapshot exists, the notice explicitly states that recorded progress is
unavailable. The rule does not project whether the target will be achieved.

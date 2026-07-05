# Fixed Deposits

## Requirement

A user can add a fixed deposit to a specific bank account with:

- name and optional bank reference
- principal amount and currency
- annual interest rate percentage
- start date and maturity date
- explicit current value and valuation date

Only accounts with type `bank` can own fixed deposits. The fixed deposit currency
must match the account currency.

## Ledger and valuation model

Creating a fixed deposit is atomic and creates:

1. a debt-class fixed-deposit asset;
2. an immutable `buy` ledger event with a principal cash outflow and one asset unit;
3. an immutable explicit price observation equal to the supplied current value;
4. fixed-deposit contract metadata linked to all three records.

The transaction ledger remains the source of truth for ownership. Portfolio
valuation uses the normal formula `ledger-derived quantity × latest explicit
price`. The annual interest rate is disclosed contract metadata only. WealthLens
does not infer compounding frequency, payout mode, penalties, tax, accrued
interest, or maturity value.

Later current-value changes must be appended as new price observations; previous
values and the opening ledger event are never mutated.

## Validation

- principal and current value must be greater than zero;
- annual interest rate must be greater than zero and no more than 100%;
- start date cannot be in the future;
- maturity date must be after start date;
- current-value date must be between start date and today;
- account must belong to the portfolio and have type `bank`;
- account and fixed-deposit currencies must match.

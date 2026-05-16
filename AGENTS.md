# AGENTS.md

## Project Identity

This project is a deterministic financial portfolio tracking and analytics platform.

The application is NOT:

- an AI product
- a trading platform
- a brokerage integration system
- a financial advisory tool
- a market prediction engine

All insights MUST be:

- formula-based
- deterministic
- explainable
- transparent
- testable

---

# Core Architecture Rules

## 1. Transaction Ledger Is Source Of Truth

Portfolio state must NEVER be stored as the primary source of truth.

Source of truth:

- transactions
- contributions
- withdrawals
- asset purchase events
- asset sale events

Derived state:

- holdings
- allocation
- portfolio value
- analytics
- risk metrics

---

## 2. Financial Logic Isolation

All financial calculations MUST live inside:

```txt
/backend/pkg/finance
```

Never place financial logic inside:

- handlers
- controllers
- repositories
- frontend components

---

## 3. Explainability Requirement

Every metric must include:

- metric name
- formula
- assumptions
- required inputs
- explanation text

No black-box logic is allowed.

---

## 4. Deterministic Outputs Only

Forbidden:

- AI-generated insights
- machine learning
- probabilistic scoring
- predictive recommendations
- hidden heuristics

Allowed:

- mathematical formulas
- benchmark comparisons
- explicit rules
- allocation thresholds
- deterministic evaluation

---

## 5. Snapshot System

Historical analytics MUST use snapshots.

Required snapshots:

- daily portfolio snapshots
- weekly performance snapshots
- monthly goal snapshots

Snapshots are append-only and immutable.

---

## 6. Backend Standards

- handlers remain thin
- services contain orchestration logic
- finance package contains calculations
- repositories only access persistence
- use dependency injection
- use explicit interfaces

---

## 7. Frontend Standards

- server components by default
- charts isolated into reusable components
- no business logic inside presentation components
- no financial formulas in frontend

---

## 8. Testing Standards

Mandatory tests:

- XIRR
- CAGR
- volatility
- drawdown
- allocation drift
- rebalancing logic
- health score

Critical financial calculations require edge-case testing.

---

## 9. Forbidden Patterns

Do NOT:

- duplicate formulas
- mutate historical snapshots
- store derived portfolio state as truth
- hardcode benchmark assumptions
- use hidden adjustment factors
- place business logic in handlers

---

## 10. Development Workflow

Before implementation:

1. plan architecture
2. identify modules
3. identify affected files
4. define edge cases
5. define validation rules

Large features must be broken into smaller vertical slices.

---

## 11. Product Philosophy

The product exists to:

- improve financial clarity
- support disciplined investing
- provide explainable analytics
- encourage long-term thinking

The product does NOT exist to:

- encourage speculation
- optimize trading
- replace financial advisors
- predict markets

---

# Architecture Document

## High-Level Architecture

The application follows a modular backend-first architecture.

The system is divided into:

1. Transaction ledger
2. Portfolio engine
3. Financial calculation engine
4. Insight engine
5. Snapshot engine
6. Visualization layer

---

# System Flow

```txt
User Transactions
        ↓
Transaction Ledger
        ↓
Holdings Engine
        ↓
Financial Calculation Engine
        ↓
Insight Engine
        ↓
Snapshot Engine
        ↓
REST API
        ↓
Frontend Dashboards
```

---

# Architectural Principles

## Source Of Truth

Transactions are immutable records.

All portfolio state is derived from transactions.

---

## Separation Of Concerns

Strict separation between:

- data storage
- calculations
- insights
- visualization

---

## Financial Engine Isolation

All formulas live in:

```txt
/backend/pkg/finance
```

The finance engine must remain framework-independent.

---

# Backend Modules

## Auth Module

Responsibilities:

- authentication
- JWT management
- password security
- session management

---

## Portfolio Module

Responsibilities:

- portfolios
- accounts
- portfolio grouping

---

## Transaction Module

Responsibilities:

- buys
- sells
- deposits
- withdrawals
- recurring investments

---

## Holdings Module

Responsibilities:

- derived holdings
- current positions
- asset aggregation

---

## Analytics Module

Responsibilities:

- PnL
- allocation
- performance calculations
- comparative analysis

---

## Risk Module

Responsibilities:

- volatility
- drawdown
- beta
- exposure analysis

---

## Goals Module

Responsibilities:

- financial goals
- projections
- contribution analysis

---

## Benchmark Module

Responsibilities:

- benchmark ingestion
- benchmark normalization
- comparative metrics

---

## Snapshot Module

Responsibilities:

- historical portfolio values
- daily snapshots
- monthly snapshots
- historical analytics

---

# Database Philosophy

## Immutable Events

Transactions should be append-only.

Avoid mutating historical transaction records.

---

## Historical Integrity

Snapshots are immutable historical records.

---

## Derived State

Portfolio values and analytics are derived state.

Never use derived state as source-of-truth data.

---

# Scalability Goals

Architecture should support:

- new asset classes
- additional benchmarks
- more risk metrics
- multiple currencies
- future mobile clients

---

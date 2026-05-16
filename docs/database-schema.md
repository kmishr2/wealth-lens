# Database Schema Overview

## Core Philosophy

Transactions are immutable.

Portfolio state is derived.

Snapshots are append-only.

---

# Core Tables

## users

Stores:

- identity
- credentials
- profile settings

---

## portfolios

Stores:

- portfolio name
- ownership
- portfolio metadata

---

## accounts

Stores:

- broker/account grouping
- bank grouping
- investment source grouping

---

## assets

Stores:

- asset identity
- asset class
- asset type
- ticker/symbol metadata

---

## transactions

Stores:

- buys
- sells
- deposits
- withdrawals
- contributions

Acts as the source of truth.

---

## holdings_snapshots

Stores:

- periodic holdings state
- historical positions

---

## portfolio_snapshots

Stores:

- historical portfolio value
- allocation history
- performance history

---

## benchmark_values

Stores:

- benchmark NAV/index values
- historical benchmark tracking

---

## goals

Stores:

- target amount
- target timeline
- expected returns

---

## reminders

Stores:

- review reminders
- contribution reminders
- drift alerts

---

# Design Rules

- UUID primary keys
- timestamps on all records
- soft delete where appropriate
- immutable financial events
- append-only historical snapshots

---

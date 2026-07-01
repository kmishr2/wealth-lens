# WealthLens - A Financial Portfolio Tracking Application

A transparent, rule-based portfolio intelligence platform for long-term investors.

## Overview

This application enables users to manually track investments and understand portfolio performance using deterministic financial calculations.

The platform is intentionally:

- AI-free
- Explainable
- Formula-driven
- Long-term investing focused

The system does NOT:

- Provide financial advice
- Predict market movements
- Use AI/LLMs
- Integrate with brokers or banks
- Perform automated trading

## Core Features

- Portfolio tracking
- Asset allocation insights
- Profit & loss analysis
- Historical performance tracking
- Benchmark comparison
- Risk indicators
- Diversification analysis
- Rebalancing suggestions
- Goals tracking
- Portfolio health scoring
- Cash flow tracking
- Scenario analysis
- Tax awareness

## Architecture Principles

- Transaction-ledger-based system
- Deterministic calculations only
- Fully explainable outputs
- Immutable historical snapshots
- Financial logic isolated from API/UI layers
- Modular financial engine

## Tech Stack

### Backend

- Go
- PostgreSQL
- Redis
- REST API

### Frontend

- Next.js
- TypeScript
- TailwindCSS
- Recharts

### Infrastructure

- Docker Compose
- GitHub Actions

## Repository Structure

```txt
/docs
/backend
/frontend
/scripts
```

## Development Philosophy

This project prioritizes:

- clarity over complexity
- transparency over automation
- explainability over intelligence
- long-term maintainability

## Status

Currently in active development.

## Setup commands for DB Testing

```bash
cp .env.example .env
make db-up
make migrate-up
make test
make run
```

In a second terminal, start the Next.js frontend:

```bash
cp frontend/.env.example frontend/.env.local
make frontend-install
make frontend-dev
```

Open `http://localhost:3000`. Authentication tokens are stored in HTTP-only
cookies and all protected API calls are made from the Next.js server.

Validate the frontend with:

```bash
make frontend-check
```

Run the PostgreSQL-backed integration tests separately:

```bash
make test-integration
```

The integration test creates a uniquely named temporary database, applies all
migrations, verifies ledger ownership and relationship constraints, runs the
real daily snapshot job twice, verifies calculated snapshots and immutability,
and drops the temporary database. It never uses the development database.

## Daily Portfolio Snapshots

Create immutable daily snapshots for every active portfolio by passing an
explicit UTC date. The command is idempotent: an existing snapshot for the
same portfolio and date is returned rather than duplicated.

```bash
make snapshot-daily DATE=2026-01-15
```

The command prints a JSON summary, continues when an individual portfolio
fails, and exits with a non-zero status if any snapshot could not be created.
Run it from cron or an external scheduler after all prices for the UTC day have
been recorded. Use the same command with a historical date for manual
backfills.

## Weekly Performance Snapshots

Create immutable weekly performance snapshots for every active portfolio by
passing the UTC Sunday week-ending date. The job compares the daily snapshot
from seven days earlier with the Sunday daily snapshot, includes external
deposit/withdrawal cash flows during the week, and stores deterministic PnL,
CAGR, and XIRR values per currency.

```bash
make snapshot-weekly DATE=2026-01-11
```

The command is idempotent: an existing weekly performance snapshot for the
same portfolio and week-ending date is returned rather than duplicated. Run it
after the Sunday daily portfolio snapshot has been created. Missing boundary
daily snapshots are reported per portfolio without stopping the remaining
portfolios.

## Nightly Indian Market Prices

The price job uses AMFI's public official daily NAV feed for Indian mutual
funds and Upstox historical candles for NSE/BSE equities. Upstox requires an
access or read-only analytics token in `UPSTOX_ACCESS_TOKEN`.

Direct automated collection from NSE's website is deliberately not used
because its published terms prohibit systematic automated collection unless
separately licensed. Provider prices remain immutable and repeated runs do not
create duplicates.

Configure each asset with a stable provider identifier. Mutual funds use the
AMFI scheme code; equities use the ISIN-based Upstox instrument key:

```sql
INSERT INTO asset_identifiers (asset_id, provider, identifier)
VALUES
  ('<mutual-fund-asset-uuid>', 'amfi', '120503'),
  ('<equity-asset-uuid>', 'upstox', 'NSE_EQ|INE002A01018');
```

Run manually (the default range ends yesterday in `Asia/Kolkata`):

```bash
make prices-nightly
make prices-nightly FROM=2026-06-25 TO=2026-06-29
```

Schedule at 00:30 IST with the deployment scheduler. For standard cron:

```cron
CRON_TZ=Asia/Kolkata
30 0 * * * cd /absolute/path/to/wealth-lens && /usr/bin/make prices-nightly >> /var/log/wealth-lens-prices.log 2>&1
```

If the portfolio contains fund-of-funds schemes, run it again around 10:30
IST because those NAVs may arrive the next business morning. The retry is safe
because ingestion is idempotent. The command prints JSON and exits non-zero
when a configured provider fails.

Pending market-data task: generate an Upstox Analytics Token and complete one
live NSE/BSE equity ingestion test before deployment.

## Benchmark Data and Comparison

Benchmarks are explicit data records, not hardcoded assumptions. Create the
benchmark first, then add immutable dated observations. Portfolio comparison
uses exact daily portfolio snapshots and exact benchmark observations for the
same start date, end date, and currency. It does not interpolate missing
benchmark values, convert currencies, or choose a default benchmark.

Example:

```bash
curl -X POST "$API_URL/api/v1/benchmarks" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "NIFTY50",
    "name": "Nifty 50",
    "currency": "INR",
    "source": "manual import",
    "description": "Indian large-cap equity benchmark"
  }'

curl -X POST "$API_URL/api/v1/benchmarks/<benchmark-id>/observations" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "observation_date": "2026-01-01",
    "value": "22000.0000",
    "source": "manual import"
  }'

curl "$API_URL/api/v1/portfolios/<portfolio-id>/benchmarks/<benchmark-id>/comparison?start_date=2026-01-01&end_date=2026-02-01&currency=INR" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

The response includes portfolio total return, benchmark total return,
portfolio CAGR, benchmark CAGR, excess return, excess CAGR, and metric metadata
explaining the formula and assumptions.

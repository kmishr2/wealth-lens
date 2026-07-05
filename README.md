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
- Bank-account fixed deposits with explicit immutable valuations
- Atomic account transaction CSV imports
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

## Full-stack containers

The default `compose.yaml` intentionally starts only PostgreSQL for local
development and integration tests. An opt-in overlay builds and starts the
migration job, Go API, and standalone Next.js frontend:

```bash
cp .env.example .env
# Replace JWT_SECRET with a unique value of at least 32 characters.
make app-up
make app-logs
```

Open `http://localhost:3000`. The API health endpoint is available at
`http://localhost:8080/healthz`. The API waits for migrations, and the frontend
waits for the API health check. Both application containers run as non-root
users.

Stop the application stack without deleting the PostgreSQL volume:

```bash
make app-down
```

This overlay is a production-like local topology, not a complete production
deployment. Replace the local database credentials, terminate TLS at a trusted
proxy, configure backups, and use the deployment scheduler for price and
snapshot jobs before exposing it publicly.

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

## Fixed Deposits

Fixed deposits can be added only to bank accounts. Creation atomically records
the contract metadata, a principal cash outflow, one fixed-deposit asset unit,
and an explicit current-value observation. Later values are appended rather
than replacing history. Annual ROI is disclosed metadata and is not used to
estimate accrued or maturity value because compounding and payout conventions
are not assumed. See `docs/fixed-deposits.md` for the complete requirements and
validation rules.

## Transaction CSV Imports

Account ledgers accept CSV files containing up to 1,000 deposits, withdrawals,
purchases, sales, fees, or taxes. The backend validates every row and writes the
entire file in one database transaction; any failure rolls back all rows. See
`docs/transaction-csv-import.md` for the exact header and validation rules.

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

## Monthly Goal Snapshots

Goals are explicit portfolio targets. Create goals through the API, then run
the monthly snapshot job after the month-end daily portfolio snapshot exists.
The job calculates current value, progress percentage, remaining amount,
months remaining, and required monthly contribution without assuming future
returns, inflation, taxes, fees, or currency conversion.

```bash
make snapshot-goals-monthly DATE=2026-01-31
```

`DATE` must be the final UTC day of a month. Monthly goal snapshots are
immutable and idempotent by goal and month-end date. A missing month-end daily
portfolio snapshot or missing goal currency is reported as a deterministic
validation failure.

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

Historical beta is available from:

```bash
curl "$API_URL/api/v1/portfolios/<portfolio-id>/benchmarks/<benchmark-id>/beta?start_date=2026-01-01&end_date=2026-06-30&currency=INR" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Beta uses exact dates shared by immutable daily portfolio snapshots and
benchmark observations. Portfolio returns remove ledger-derived deposits and
withdrawals. At least three aligned observations and non-zero benchmark return
variance are required; no interpolation or currency conversion is performed.

## Asset Concentration

`GET /api/v1/portfolios/<portfolio-id>/concentration` calculates asset
concentration separately per currency from current ledger-derived holdings and
explicit market prices. It reports HHI, effective asset count, and the largest
asset percentage. Cash, currency conversion, subjective labels, and advisory
thresholds are excluded. A complete valuation is required.

`GET /api/v1/portfolios/<portfolio-id>/diversification-alerts` applies the
disclosed health-score diversification bands to the same current valuation.
It returns `none`, `notice`, `warning`, or `critical` based on the lower of the
largest-asset and holding-count band scores, together with every triggered
condition. These are factual threshold alerts, not investment recommendations.

## Portfolio Health Score

The health score is calculated separately per currency from five disclosed
components: diversification (25), allocation drift (25), volatility (20),
maximum drawdown (15), and data quality (15). It uses current ledger-derived
allocation plus the trailing 12 months of immutable daily snapshots, annualized
at 252 periods per year. The default risk profile is `moderate`.

```bash
curl -X POST "$API_URL/api/v1/portfolios/<portfolio-id>/health-score" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"as_of_date":"2026-06-30","risk_profile":"moderate"}'
```

Available defaults are conservative (30/60/10 allocation, 8% volatility,
-10% drawdown), moderate (60/35/5, 15%, -20%), and aggressive (85/10/5,
25%, -35%). Allocation values represent equity/debt/cash-other percentages.
Per-currency custom targets and risk thresholds can be supplied in
`currency_configurations`; all three targets must be present and sum to 100.

Assets now expose an optional `risk_category`: `equity`, `debt`, or
`cash_other`. Equity, bond, and cash asset classes receive deterministic
defaults. Funds and other ambiguous asset classes must be classified explicitly
or receive the partial-data-quality score.

## SIP Projections

SIP projections are stateless deterministic scenarios. The caller supplies the
currency, initial investment, monthly contribution, nominal annual return,
optional non-negative inflation, and a horizon from 1 to 1200 months.

```bash
curl -X POST "$API_URL/api/v1/portfolios/<portfolio-id>/projections/sip" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "currency": "INR",
    "initial_investment": "100000",
    "monthly_contribution": "10000",
    "annual_return_percentage": "10",
    "annual_inflation_percentage": "6",
    "months": 120
  }'
```

Contributions are applied at each month end and rates compound monthly. The
response includes every monthly point, total contributions, nominal growth,
and nominal and inflation-adjusted values. Assumed returns remain explicit and
are simulations, not predictions. Projected values are never stored as
portfolio truth.

Compare two to ten named scenarios with the same currency and horizon using
`POST /api/v1/portfolios/<portfolio-id>/projections/what-if`. The first
scenario is the explicit baseline. Each scenario contains an `input` object
with the same amount and rate fields as a SIP projection, excluding currency.
The response includes each full projection and its nominal value, real value,
and contribution differences from the baseline. Scenario order and results are
deterministic, and nothing is persisted.

## Contribution Analysis

Historical contribution attribution uses exact boundary snapshots and the
immutable transaction ledger:

```bash
curl "$API_URL/api/v1/portfolios/<portfolio-id>/contributions?start_date=2026-01-31&end_date=2026-06-30&currency=INR" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

The response separates positive deposits, withdrawals, net contributions, and
investment growth using `ending value - beginning value - net contributions`.
It also returns UTC monthly cash-flow buckets. Both boundary snapshots must be
fully valued and contain the requested currency; no interpolation or currency
conversion is performed.

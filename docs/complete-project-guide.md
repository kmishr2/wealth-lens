# WealthLens Complete Project Guide

This document describes the implemented WealthLens system as of 7 July 2026.
It is intended to let a developer, tester, reviewer, or operator understand the
product scope, architecture, database, application flows, implemented features,
and validation procedures without first reading the entire repository.

The source of truth for this guide is the current application code and database
migrations. Where an older conceptual document differs from the implementation,
this guide describes the implementation.

## 1. Product purpose and boundaries

WealthLens is a deterministic financial portfolio tracking and analytics
application for long-term investors. Users manually record financial events or
import them from CSV, attach explicit prices, and receive formula-based views of
holdings, value, allocation, performance, risk, goals, and data quality.

The product is deliberately not:

- a brokerage or bank integration;
- an order execution or trading platform;
- an investment advisory service;
- a market prediction product;
- an AI or machine-learning system;
- a foreign-exchange conversion engine;
- a tax calculation engine.

Every calculation is deterministic. Financial responses include a metric
definition containing the formula, assumptions, required inputs, and an
explanation. Missing information is disclosed instead of estimated.

## 2. Implemented scope at a glance

The current implementation includes:

- registration, login, rotating refresh sessions, logout, and profile settings;
- portfolios and portfolio-scoped accounts;
- an asset catalogue with asset classes and risk categories;
- an immutable transaction ledger with reversals and corrections;
- atomic transaction CSV imports;
- ledger-derived holdings and cash balances;
- immutable manual and provider price observations;
- nightly Indian mutual-fund and equity price ingestion;
- per-currency valuation and allocation;
- concentration, diversification, drift, and rebalancing calculations;
- immutable daily portfolio, weekly performance, and monthly goal snapshots;
- PnL, CAGR, XIRR, volatility, maximum drawdown, and contribution attribution;
- benchmark series, comparison, and historical beta;
- deterministic portfolio health scoring;
- goals, SIP projections, and what-if scenarios;
- bank-account fixed deposits with maturity and closure workflows;
- deterministic maturity, goal-date, missing-price, and stale-price notices;
- ledger and portfolio-summary CSV exports;
- local full-stack container builds and PostgreSQL integration tests.

## 3. Technology and repository layout

| Layer | Technology | Location |
| --- | --- | --- |
| Backend API and jobs | Go 1.23, Gin, GORM | `backend/` |
| Financial engine | Go, `shopspring/decimal` | `backend/pkg/finance/` |
| Database | PostgreSQL 16 | `backend/migrations/` |
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS | `frontend/` |
| Local orchestration | Docker Compose | `compose.yaml`, `compose.app.yaml` |
| Developer commands | Make | `Makefile` |
| Integration harness | Shell, Docker, golang-migrate | `scripts/` |
| Product and engineering documentation | Markdown | `docs/` |

The Go backend uses module packages under `backend/internal`. HTTP handlers,
services, repositories, and DTOs are kept separate. Framework-independent
financial formulas live only in `backend/pkg/finance`.

## 4. High-level architecture

```text
Browser
  |
  | HTTPS in a real deployment
  v
Next.js frontend
  |  server-side API calls with access token from HTTP-only cookie
  v
Gin REST API (/api/v1)
  |
  +--> handlers: HTTP parsing and response mapping
  +--> services: ownership checks, validation, orchestration
  +--> finance package: deterministic calculations
  +--> repositories: persistence queries only
  v
PostgreSQL
  +--> immutable ledger and observations
  +--> append-only snapshots
  +--> mutable configuration/profile records

External scheduled commands
  +--> AMFI and Upstox price job
  +--> daily portfolio snapshot job
  +--> weekly performance snapshot job
  +--> monthly goal snapshot job
```

### 4.1 Layer responsibilities

- **Frontend server components** load data and render pages. Client components
  collect inputs. Financial formulas are not implemented in the frontend.
- **Handlers** parse authenticated route parameters, query strings, JSON, or CSV
  and map application errors to the standard API envelope.
- **Services** enforce ownership, validation, idempotency, and workflow rules.
- **Finance functions** perform calculations using explicit typed inputs and
  return typed results with metric definitions.
- **Repositories** read and write PostgreSQL without embedding financial rules.
- **Database constraints and triggers** enforce cross-tenant relationships,
  uniqueness, permitted states, and immutability.

### 4.2 Core data flow

```text
User input
  -> transaction + signed entries
  -> immutable ledger
  -> sum entries to derive holdings and cash
  -> combine holdings with latest explicit prices
  -> calculate per-currency value and allocation
  -> calculate analytics or append immutable snapshots
  -> expose result and metric definition through API
  -> render UI or export CSV
```

No current holdings table exists. Holdings, cash, current value, and current
allocation are calculated from the ledger whenever requested.

## 5. Architectural rules implemented by the system

### 5.1 Ledger source of truth

`transactions` and `transaction_entries` are the source of truth for ownership
and cash. A purchase adds a positive asset quantity and normally a negative cash
entry. A sale adds a negative asset quantity and positive cash. Deposits,
withdrawals, fees, and taxes use signed amount entries.

Historical transactions are not edited or deleted. A reversal appends entries
with opposite signs. A correction atomically appends a reversal and a corrected
replacement linked by a correction group.

### 5.2 Per-currency calculations

The application does not silently convert currencies. Holdings and results are
grouped by their three-letter currency. A portfolio may therefore have several
totals and several allocation groups. Cross-currency comparison requires data
already expressed in the requested currency; no exchange-rate assumption is
made.

### 5.3 Explicit observations

Prices and benchmark values are immutable observations with dates and sources.
If a required price is missing, valuation and allocation return a missing-price
record and mark the result incomplete. They do not use a synthetic value.

### 5.4 Historical snapshots

Historical analytics use immutable snapshots:

- daily portfolio snapshots;
- Sunday-ending weekly performance snapshots;
- month-end goal snapshots.

Jobs are idempotent for their natural key. Repeating a successful job does not
create a duplicate snapshot.

### 5.5 Explainability

Metrics expose `name`, `formula`, `assumptions`, `required_inputs`, and
`explanation`. Portfolio summary exports include these definitions as separate
rows so the exported values remain auditable outside the application.

## 6. Backend module map

| Module | Responsibility |
| --- | --- |
| `auth` | Registration, login, JWT access tokens, hashed refresh sessions, rotation, logout |
| `users` | Current-user profile retrieval and update |
| `portfolios` | Portfolio lifecycle and ownership |
| `accounts` | Portfolio-scoped brokerage, retirement, bank, wallet, and other accounts |
| `assets` | Asset reference data, classes, currencies, exchanges, risk categories |
| `transactions` | Ledger creation, listing, CSV import, reversal, correction |
| `holdings` | Ledger query and holdings calculation |
| `prices` | Manual price observations and automated ingestion persistence |
| `marketdata` | AMFI and Upstox provider clients and nightly price orchestration |
| `valuations` | Current per-currency value and missing-price disclosure |
| `allocations` | Asset/class/cash allocation, concentration, diversification, rebalancing |
| `snapshots` | Daily snapshots, weekly job orchestration, snapshot listing |
| `performance` | Snapshot-backed PnL, CAGR, and XIRR |
| `risk` | Cash-flow-adjusted volatility and maximum drawdown |
| `benchmarks` | Benchmark definitions, observations, and portfolio comparison |
| `beta` | Cash-flow-adjusted historical beta against a benchmark |
| `health` | Five-component deterministic health score |
| `goals` | Goal lifecycle and immutable monthly progress snapshots |
| `projections` | Stateless SIP and what-if calculations |
| `contributions` | Deposit/withdrawal attribution and investment growth |
| `fixeddeposits` | Bank-account fixed-deposit creation, values, maturity, and closure |
| `notifications` | Derived fixed-deposit, goal-date, and market-data notices |
| `common` | API envelopes, errors, currency helpers, pagination |
| `middleware` | Authentication, request IDs, and auth endpoint rate limiting |
| `config` | Environment loading and production validation |
| `server` | Dependency construction and route registration |

## 7. Authentication and authorization flow

1. Registration validates the profile and hashes the password with bcrypt.
2. Login verifies the password and issues a short-lived JWT access token plus a
   random refresh token.
3. Only the SHA-256 refresh-token hash is stored in `auth_sessions`.
4. The frontend stores both tokens in HTTP-only, `SameSite=Lax` cookies. Cookies
   are marked `Secure` in production.
5. Protected frontend pages call the backend from the Next.js server and attach
   the JWT as a bearer token.
6. Refresh atomically consumes the active session and creates a replacement,
   preventing reuse of the old refresh token.
7. Logout revokes the refresh session and clears cookies.
8. Every portfolio resource is checked against the authenticated owner. The
   database also has composite foreign keys preventing cross-portfolio account,
   creator, reversal, correction, snapshot, and fixed-deposit relationships.

Auth routes are rate-limited. Production configuration rejects the development
JWT secret and secrets shorter than 32 characters.

## 8. Implemented features and how to test them

The API examples below assume:

```bash
export API_URL=http://localhost:8080/api/v1
export ACCESS_TOKEN='<token returned by login>'
```

The frontend is available at `http://localhost:3000` during local development.

### 8.1 User registration, login, refresh, logout, and profile

**What it does:** creates users, authenticates passwords, rotates refresh
sessions, and manages display name, base currency, and timezone.

**UI test:** register at `/register`, sign in at `/login`, open `/profile`, edit
the profile, refresh the browser, sign out, and confirm protected pages redirect
to login.

**API test:** call `POST /auth/register`, `POST /auth/login`,
`POST /auth/refresh`, `GET /users/me`, `PATCH /users/me`, and
`POST /auth/logout`. Confirm an already-consumed refresh token cannot be reused.

**Automated coverage:** `internal/auth`, `internal/middleware`, and
`internal/config` unit/integration tests.

### 8.2 Portfolios and accounts

**What it does:** groups all user data into owned portfolios and accounts.
Portfolio and account names are unique within their active ownership scope.
Supported account types are brokerage, retirement, bank, wallet, and other.

**UI test:** create a portfolio on `/dashboard`; open it; add, edit, and delete an
empty account; verify account currency and type appear correctly.

**API test:** use portfolio CRUD endpoints and nested account CRUD endpoints.
Attempt to access an identifier owned by another user and expect `404` rather
than data leakage.

### 8.3 Asset catalogue and risk classification

**What it does:** stores reusable asset identity, symbol, exchange, asset class,
currency, active status, and optional risk category. Asset classes include cash,
equity, fund, bond, fixed deposit, crypto, real estate, commodity, alternative,
and other. Risk categories are equity, debt, or cash/other.

**UI test:** create an asset at `/assets`, open its detail page, and record a
manual price. Fixed-deposit assets do not appear in the generic catalogue
because their lifecycle is controlled by the fixed-deposit workflow.

**API test:** create and list assets, then `GET /assets/:assetId`. Verify symbols
and exchanges normalize to uppercase and duplicate symbol/exchange/currency
combinations are rejected.

### 8.4 Immutable transaction ledger

**What it does:** records deposits, withdrawals, buys, sells, fees, taxes,
transfers, reversals, and corrections as transactions with signed entries.
Idempotency keys are unique per portfolio when supplied.

**UI test:** open an account and record a deposit, buy, fee, and sale. Confirm
the event list shows entry signs. Reverse an ordinary event, then correct another
one; verify the old rows still exist and linked audit rows were appended.

**API test:** create a transaction with `POST /portfolios/:id/transactions`,
list/get it, reverse it through `/reversals`, and correct it through
`/corrections`. Database attempts to update or delete ledger rows must fail.

**Important rule:** fixed-deposit assets and their opening/closure transactions
cannot be changed through generic transaction, import, reversal, or correction
endpoints.

### 8.5 Atomic transaction CSV import

**What it does:** imports up to 1,000 account transactions from a file no larger
than 2 MiB. Every row is validated before any transaction is committed. One bad
row rolls back the entire import.

Required header:

```csv
transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key
```

**UI test:** use the import section on an account page. Import a valid file and
verify every row appears. Then import a file with one invalid row and verify none
of its rows were added.

**Automated coverage:** `backend/internal/transactions/csv_import_test.go` and
the PostgreSQL rollback test in `integrity_integration_test.go`.

### 8.6 Holdings and cash

**What it does:** sums signed asset quantities by asset and signed cash, fee, and
tax amounts by currency. Zero-result positions are removed.

Formula:

```text
asset quantity = sum(signed asset quantities)
cash balance = sum(signed cash + fee + tax amounts)
```

**UI test:** add a deposit and buy, then view the portfolio overview. Add a sale
or reversal and verify holdings change without editing any stored position.

**API test:** `GET /portfolios/:id/holdings`.

### 8.7 Manual asset prices

**What it does:** appends positive, dated, source-labelled price observations.
Observations cannot be updated or deleted.

**UI test:** open `/assets/:assetId`, add two prices on different dates, and
confirm the history and latest price are displayed.

**API test:** use `POST /assets/:assetId/prices`, list prices, and request
`/latest`. Try a non-positive price, currency mismatch, or future date and expect
validation failure.

### 8.8 Nightly Indian market prices

**What it does:** fetches official AMFI NAV data for Indian mutual funds and
Upstox historical candles for NSE/BSE equities. Provider identifiers are stored
in `asset_identifiers`. Automated observations are idempotent by asset, market
date, and source.

Identifiers:

- AMFI: scheme code, for example `120503`;
- Upstox: instrument key, for example `NSE_EQ|INE002A01018`.

**Manual test:** configure identifiers, set `UPSTOX_ACCESS_TOKEN` when equities
are present, and run:

```bash
make prices-nightly FROM=2026-06-25 TO=2026-06-29
```

Run it twice and verify the second run creates no duplicate provider prices.
AMFI and Upstox HTTP parsing is covered with deterministic test servers. A live
Upstox ingestion test remains a deployment task because it needs a real token.

### 8.9 Current valuation

**What it does:** multiplies holdings by latest explicit prices and adds cash,
separately per currency. Missing prices are returned explicitly.

```text
asset market value = quantity × explicit price
currency total = sum(asset market values) + cash balance
```

**UI test:** view a portfolio containing one priced and one unpriced holding.
Confirm the unpriced asset is named and the valuation is marked incomplete.
Record its price and verify completeness becomes true.

**API test:** `GET /portfolios/:id/valuation`.

### 8.10 Allocation, concentration, and diversification alerts

**What it does:** calculates asset, asset-class, and cash percentages per
currency. Concentration reports HHI, effective asset count, largest position,
and holding count. Diversification alerts apply disclosed bands.

```text
allocation % = component value / same-currency total × 100
HHI = sum((asset weight × 100)^2)
effective asset count = 10000 / HHI
```

Diversification scoring uses the lower of the largest-position and holding-count
bands: 25 points for the strongest band, then 18, 10, or 0. Severities are none,
notice, warning, and critical.

**UI test:** inspect allocation on the portfolio and analytics pages. Add or
remove positions and confirm percentages and concentration change.

**API test:** call `/allocation`, `/concentration`, and
`/diversification-alerts`. Incomplete valuation must not produce a misleading
complete concentration result.

### 8.11 Rebalancing calculation

**What it does:** compares current asset-class allocation with explicit targets
and a supplied tolerance. It returns drift and value adjustments; it does not
place trades or account for tax, fees, liquidity, or suitability.

```text
drift = current percentage - target percentage
target value = currency total × target percentage / 100
suggested adjustment = target value - current value
```

**UI test:** use the rebalancing form on the analytics page, supply targets that
sum to 100 per currency, and compare results above and within tolerance.

**API test:** `POST /portfolios/:id/rebalancing`.

### 8.12 Daily portfolio snapshots

**What it does:** appends a complete historical record of values, allocation,
missing prices, scopes, and metric metadata for a UTC date. The same
portfolio/date/period is idempotent and immutable.

**Manual test:**

```bash
make snapshot-daily DATE=2026-01-15
make snapshot-daily DATE=2026-01-15
```

Confirm only one row exists and `GET /portfolios/:id/snapshots` returns it.
Attempts to update or delete it must fail.

### 8.13 Performance: PnL, CAGR, and XIRR

**What it does:** uses immutable boundary snapshots and external cash flows to
measure observed performance separately by currency.

```text
period PnL = ending value - beginning value - net external cash flow
CAGR = ((ending / beginning)^(365.25 / elapsed days) - 1) × 100
XIRR solves: sum(cash flow / (1+r)^(days/365.25)) = 0
```

**UI test:** create boundary daily snapshots with intervening deposits or
withdrawals, then inspect the analytics page.

**API test:** `GET /portfolios/:id/performance` with the required dates. The
response contains a result for each available currency. Missing exact boundary
snapshots must return an explicit error.

### 8.14 Weekly performance snapshots

**What it does:** creates an immutable Sunday-ending record from exact daily
snapshots seven days apart, including per-currency PnL, CAGR, XIRR, and metadata.

**Manual test:** create daily snapshots for both Sundays, record any intervening
cash flows, then run:

```bash
make snapshot-weekly DATE=2026-01-11
```

List results through `/weekly-performance-snapshots`. Repeat the job to verify
idempotency. Non-Sunday dates and missing boundaries must fail deterministically.

### 8.15 Historical risk

**What it does:** calculates annualized sample volatility and maximum drawdown
from snapshots, removing external contributions and withdrawals from periodic
returns. The standard annualization frequency is explicit in the request/service
and health scoring uses 252 periods per year.

```text
volatility = sample standard deviation(periodic returns) × sqrt(periods/year) × 100
adjusted return = (current value - external cash flow) / previous value - 1
drawdown = wealth index / prior running peak - 1
```

**UI test:** inspect risk on the analytics page after enough daily snapshots
exist.

**API test:** `GET /portfolios/:id/risk` with start date, end date, and
`periods_per_year`. The response is grouped by currency. Test constant returns,
deposits, losses, and insufficient history.

### 8.16 Benchmarks and comparison

**What it does:** stores benchmark definitions and immutable dated observations.
Comparison requires exact matching start and end dates and the same currency; no
default benchmark or interpolation is used.

**UI test:** create a benchmark at `/benchmarks`, add observations, then select
it on a portfolio analytics page.

**API test:** create/list benchmarks and observations, then call
`GET /portfolios/:portfolioId/benchmarks/:benchmarkId/comparison` with exact
dates and currency. Verify total return, CAGR, excess return, and excess CAGR.

### 8.17 Historical beta

**What it does:** calculates covariance of aligned, cash-flow-adjusted portfolio
returns with benchmark returns divided by benchmark variance. At least three
aligned observations and non-zero benchmark variance are required.

**API test:** call `/portfolios/:portfolioId/benchmarks/:benchmarkId/beta` over
a range with aligned daily snapshots and observations. Test a zero-variance
benchmark and insufficient overlap.

### 8.18 Contribution analysis

**What it does:** separates deposits, withdrawals, net contributions, and
investment growth between exact boundary snapshots. It also returns monthly cash
flow buckets.

```text
net contributions = contributions - withdrawals
investment growth = ending value - beginning value - net contributions
```

**UI test:** inspect the contribution section on analytics after recording
deposits and withdrawals between two snapshots.

**API test:** `GET /portfolios/:id/contributions` with exact dates and currency.

### 8.19 Portfolio health score

**What it does:** adds five disclosed components per currency:

| Component | Maximum points |
| --- | ---: |
| Diversification | 25 |
| Allocation drift | 25 |
| Volatility | 20 |
| Maximum drawdown | 15 |
| Data quality | 15 |

Default profiles are conservative (30/60/10 equity/debt/cash-other, 8%
volatility, -10% drawdown), moderate (60/35/5, 15%, -20%), and aggressive
(85/10/5, 25%, -35%). Callers may provide per-currency custom targets and risk
thresholds. Every observed value, threshold, band, and awarded point is returned.

The implemented bands are:

| Component condition | Points |
| --- | ---: |
| Largest asset ≤20% and at least 8 holdings | 25 |
| Largest asset >20–35% or 5–7 holdings | 18 |
| Largest asset >35–50% or 3–4 holdings | 10 |
| Largest asset >50% or fewer than 3 holdings | 0 |
| Allocation drift 0–5% / 5–10% / 10–20% / 20–30% / >30% | 25 / 20 / 12 / 6 / 0 |
| Volatility ≤80% / 80–100% / 100–125% / 125–150% / >150% of threshold | 20 / 16 / 10 / 5 / 0 |
| Drawdown ≤50% / 50–100% / 100–125% / 125–150% / >150% of threshold | 15 / 12 / 8 / 4 / 0 |
| Data quality complete / minor gaps / missing class or history / major gaps | 15 / 10 / 5 / 0 |

Diversification takes the lower of the largest-asset and holding-count scores.

**UI test:** calculate health on the analytics page using each profile and custom
targets. Change an asset risk category or add missing data and verify the data
quality/drift components change.

**API test:** `POST /portfolios/:id/health-score`.

### 8.20 Goals and monthly goal snapshots

**What it does:** stores explicit target amount, currency, date, and status. A
month-end snapshot compares the exact portfolio currency value with the target.
It does not assume returns or forecast success.

```text
progress % = current value / target value × 100
remaining = max(target - current, 0)
required monthly contribution = remaining / calendar months remaining
```

**UI test:** create/edit/archive a goal on `/portfolios/:id/planning`. Create the
month-end daily snapshot, run the goal job, and inspect progress history.

**Manual job test:**

```bash
make snapshot-goals-monthly DATE=2026-01-31
```

The date must be the final UTC day of a month. Repeat the command to verify
idempotency.

### 8.21 SIP projections and what-if scenarios

**What it does:** runs stateless simulations from explicit initial investment,
monthly contribution, annual return, inflation, and horizon inputs. No projection
is stored as portfolio truth.

```text
monthly rate = annual percentage / 12 / 100
balance = previous balance × (1 + monthly rate) + end-of-month contribution
real value = nominal value / (1 + monthly inflation rate)^months
```

What-if comparison accepts two to ten uniquely named scenarios and reports each
scenario's difference from the first explicit baseline.

**UI test:** use the calculators on the planning page with zero, positive, and
negative assumed returns and with inflation.

**API test:** `POST /projections/sip` and `POST /projections/what-if`.

### 8.22 Fixed deposits

**What it does:** adds a fixed deposit only to a bank account. Creation atomically
creates a debt-class asset, a buy ledger event with principal cash outflow, one
asset unit, an explicit value observation, and contract metadata.

Annual ROI is descriptive metadata. WealthLens does not infer compounding,
interest payout, penalties, tax, accrued value, or maturity value.

Later values append immutable observations. Status is derived as active,
maturity due, or closed. Closure atomically adds a sell event removing one unit,
adds actual cash proceeds, and stores immutable maturity or premature-closure
metadata. Closed deposits cannot receive values or close twice.

**UI test:** on a bank account, create an FD, append a value, observe maturity
status, and close it. Confirm holdings drop by one unit and cash increases by the
actual proceeds. Try generic reversal or correction and expect rejection.

**Automated coverage:** unit tests plus a PostgreSQL integration test of creation,
ledger protection, closure, and notification removal.

### 8.23 Deterministic notices

**What it does:** derives current notices without storing a separate reminder
truth:

- open fixed deposits within 30 days of maturity or overdue;
- active unreached goals within 30 days of target date or overdue;
- held non-FD assets with no price;
- held non-FD assets whose price age exceeds a disclosed threshold.

Date statuses are upcoming (8–30 days), urgent (1–7), due (today), and overdue.
The stale-price default is five calendar days and can be overridden from 1 to
365 using `stale_after_days`.

**UI test:** open `/notifications` with qualifying records. Close an FD, complete
a goal, or add a price and confirm its notice disappears.

**API test:** `GET /notifications?as_of_date=2026-07-07&stale_after_days=5`.
Historical requests exclude future transactions and price observations.

### 8.24 Ledger export

**What it does:** downloads portfolio or account transaction entries as a CSV,
including transaction links, account names, asset symbols, quantities, amounts,
currencies, and audit references. It paginates through all data and disables
caching.

**UI test:** click **Export ledger** on a portfolio or account page. Open the CSV
and compare row counts and identifiers with the ledger.

### 8.25 Portfolio summary export

**What it does:** downloads a long-form CSV containing per-currency totals,
holdings, cash, allocation, missing-price disclosures, and metric definitions.
It does not combine currencies or infer missing values. Text is protected from
spreadsheet formula injection.

**UI test:** click **Export summary** on a portfolio page. Verify the row types
`portfolio_total`, `asset_holding`, `cash_balance`,
`asset_class_allocation`, `missing_price`, and `metric_definition`.

**Automated coverage:** `frontend/test/portfolio-report.test.ts` verifies
multi-currency separation, explainability rows, and spreadsheet safety.

## 9. REST API inventory

All backend endpoints use `/api/v1`. Except registration, login, refresh, and
logout, routes require a bearer access token.

| Area | Methods and paths |
| --- | --- |
| Auth | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout` |
| User | `GET/PATCH /users/me` |
| Portfolios | `POST/GET /portfolios`, `GET/PATCH/DELETE /portfolios/:portfolioId` |
| Accounts | `POST/GET /portfolios/:portfolioId/accounts`, `GET/PATCH/DELETE .../accounts/:accountId` |
| Assets | `POST/GET /assets`, `GET /assets/:assetId` |
| Prices | `POST/GET /assets/:assetId/prices`, `GET .../prices/latest` |
| Transactions | `POST/GET /portfolios/:portfolioId/transactions`, `GET .../transactions/:transactionId` |
| Audit workflows | `POST .../:transactionId/reversals`, `POST .../:transactionId/corrections` |
| CSV import | `POST /portfolios/:portfolioId/accounts/:accountId/transaction-imports` |
| Holdings/value | `GET /portfolios/:portfolioId/holdings`, `/valuation`, `/allocation` |
| Allocation analytics | `GET .../concentration`, `GET .../diversification-alerts`, `POST .../rebalancing` |
| Snapshots | `POST/GET .../snapshots`, `GET .../weekly-performance-snapshots` |
| Analytics | `GET .../performance`, `/risk`, `/contributions` |
| Benchmarks | `POST/GET /benchmarks`, `POST/GET /benchmarks/:id/observations` |
| Benchmark analytics | `GET .../benchmarks/:id/comparison`, `GET .../benchmarks/:id/beta` |
| Health | `POST /portfolios/:portfolioId/health-score` |
| Goals | `POST/GET .../goals`, `PATCH/DELETE .../goals/:goalId` |
| Goal snapshots | `POST/GET .../goals/:goalId/monthly-snapshots` |
| Projections | `POST .../projections/sip`, `POST .../projections/what-if` |
| Fixed deposits | `POST/GET .../accounts/:accountId/fixed-deposits` |
| FD values/closure | `POST .../fixed-deposits/:id/values`, `POST .../:id/closure` |
| Notices | `GET /notifications` |

Success envelope:

```json
{"success": true, "data": {}}
```

Error envelope:

```json
{"success": false, "error": {"code": "VALIDATION_ERROR", "message": "..."}}
```

## 10. Frontend route map

| Route | Purpose |
| --- | --- |
| `/register`, `/login` | Authentication |
| `/dashboard` | Portfolio list and creation |
| `/profile` | User settings |
| `/assets` | Asset catalogue and creation |
| `/assets/:assetId` | Asset details and immutable prices |
| `/benchmarks` | Benchmark list and creation |
| `/benchmarks/:benchmarkId` | Observation history and creation |
| `/portfolios/:portfolioId` | Accounts, holdings, value, allocation, exports |
| `/portfolios/:portfolioId/accounts/:accountId` | Ledger, import, transaction entry, FDs |
| `/portfolios/:portfolioId/analytics` | Performance, risk, concentration, health, benchmark analytics |
| `/portfolios/:portfolioId/planning` | Goals, goal history, SIP, what-if |
| `/notifications` | Derived operational and date notices |
| `/portfolios/:id/exports/transactions` | Ledger CSV response |
| `/portfolios/:id/exports/summary` | Portfolio summary CSV response |

Server components are used by default. Interactive forms are isolated client
components backed by server actions. Authentication tokens are never exposed to
browser JavaScript.

## 11. Database design

### 11.1 Relationship overview

```text
users
  +--< auth_sessions
  +--< portfolios
         +--< accounts
         +--< transactions >-- accounts
         |      +--< transaction_entries >-- assets
         +--< portfolio_snapshots
         +--< weekly_performance_snapshots
         +--< goals
         |      +--< monthly_goal_snapshots
         +--< fixed_deposits >-- accounts/assets/transactions
                +--0..1 fixed_deposit_closures >-- transactions

assets
  +--< asset_prices
  +--< asset_identifiers

benchmarks
  +--< benchmark_observations
```

Notices, current holdings, current allocation, current valuation, risk results,
health results, projections, and contribution analysis are derived and therefore
do not have primary-state tables.

### 11.2 Tables

#### `users`

Identity and preferences: UUID, case-insensitive unique email, password hash,
display name, base currency, timezone, timestamps, and soft deletion.

#### `auth_sessions`

Hashed refresh sessions: user, unique token hash, expiry, revocation timestamp,
and creation timestamp. Sessions cascade when their user is deleted.

#### `portfolios`

User-owned portfolio metadata: name, description, base currency, timestamps,
and soft deletion. Active names are unique per user.

#### `accounts`

Portfolio-owned account metadata: name, type, institution, currency, timestamps,
and soft deletion. Active names are unique per portfolio.

#### `assets`

Global asset reference data: uppercase symbol, name, class, optional risk
category, currency, uppercase exchange, active flag, and timestamps. The
symbol/exchange/currency tuple is unique.

#### `transactions`

Immutable event header: portfolio, account, type, occurrence time, description,
optional idempotency key, reversal/correction references, correction group,
creator, and timestamps. Composite constraints ensure account and audit targets
belong to the same portfolio scope. One reversal and one correction are allowed
per target.

#### `transaction_entries`

Immutable signed legs: transaction, kind, optional asset, quantity, amount,
currency, and creation time. Asset entries require asset and non-zero quantity;
non-asset entries require a non-zero amount.

#### `asset_prices`

Immutable price observations: asset, positive price, currency, timestamp, source,
note, optional creator, provider market date, and creation time. Manual records
must have a creator. Automated records are unique by asset/market-date/source.

#### `asset_identifiers`

Maps an asset to an external provider identifier. Asset/provider and
provider/identifier are both unique.

#### `portfolio_snapshots`

Immutable daily historical state: date, period, JSONB currency totals, asset,
class, and cash allocations, missing prices, completeness, scopes, three metric
definitions, creator, and creation time. Portfolio/date/period is unique.

#### `weekly_performance_snapshots`

Immutable week boundaries, JSONB per-currency returns, performance scope, PnL,
CAGR, and XIRR metadata, creator, and creation time. Week-end is unique per
portfolio.

#### `benchmarks`

Benchmark reference data: globally unique case-insensitive code, name, currency,
source, description, optional creator, and creation time.

#### `benchmark_observations`

Immutable dated positive values with source, note, creator, and creation time.
Benchmark/date is unique.

#### `goals`

Portfolio goal configuration: name, positive target amount, currency, target
date, active/completed/archived status, creator, timestamps, and soft deletion.
Active names are unique per portfolio.

#### `monthly_goal_snapshots`

Immutable month-end goal facts: current/target values, progress, remaining
amount, months remaining, required monthly contribution, reached flag, metric
metadata, creator, and creation time. Goal/month-end is unique.

#### `fixed_deposits`

Contract metadata linked one-to-one to a generated fixed-deposit asset and
opening transaction: portfolio/account, name, bank reference, principal,
currency, annual rate, start/maturity dates, creator, and creation time.

#### `fixed_deposit_closures`

Optional immutable one-to-one closure: FD scope, closing transaction, maturity or
premature type, closure date, actual proceeds, currency, note, creator, and
creation time. Composite keys guarantee FD, account, portfolio, currency, and
transaction agree.

### 11.3 Immutability enforcement

PostgreSQL `BEFORE UPDATE OR DELETE` triggers reject changes to:

- transactions;
- transaction entries;
- asset prices;
- portfolio snapshots;
- benchmark observations;
- weekly performance snapshots;
- monthly goal snapshots;
- fixed-deposit closures.

Corrections are append-only ledger events rather than mutations.

### 11.4 Soft deletion versus restriction

Users, portfolios, accounts, and goals have soft-delete fields. Financial events
use restrictive foreign keys to preserve audit history; soft deletion hides
configuration records without deleting their ledger or snapshot rows.

### 11.5 Precision and JSONB

- Monetary ledger amounts and FD principal/proceeds use `numeric(28,4)`.
- Quantities and calculated/observed values generally use `numeric(28,10)`.
- Rates use explicit decimal types; Go calculations avoid binary floating point
  except numerical root/statistical operations that are converted back with
  controlled precision.
- Snapshot aggregates and metric definitions are stored in JSONB so each
  historical result retains its exact explanatory context.

## 12. Scheduled jobs and operating order

The commands are external-scheduler friendly. No scheduler is embedded in the
API process.

Recommended order:

```text
00:30 Asia/Kolkata  -> prices-nightly
after price success -> snapshot-daily for the completed UTC date
Sunday after daily  -> snapshot-weekly
month end after daily -> snapshot-goals-monthly
```

Commands:

```bash
make prices-nightly
make snapshot-daily DATE=YYYY-MM-DD
make snapshot-weekly DATE=YYYY-MM-DD
make snapshot-goals-monthly DATE=YYYY-MM-DD
```

Price and snapshot commands print JSON summaries and exit non-zero if work
failed. Portfolio-scoped job failures are reported without hiding other
portfolio results.

## 13. Local setup and complete test procedure

### 13.1 Prerequisites

- Go 1.23 or compatible toolchain;
- Node.js 22 and npm;
- Docker with Compose;
- `golang-migrate` CLI;
- Make.

### 13.2 Start the development system

```bash
cp .env.example .env
cp frontend/.env.example frontend/.env.local
make db-up
make migrate-up
make run
```

In a second terminal:

```bash
make frontend-install
make frontend-dev
```

Open `http://localhost:3000`.

### 13.3 Backend unit tests

```bash
make test
```

This covers services and all critical finance formulas, including holdings,
valuation, allocation, concentration, diversification, rebalancing, PnL, CAGR,
XIRR, volatility, drawdown, beta, contributions, health score, goals, SIP,
what-if, fixed-deposit metadata, providers, notices, auth, and middleware.

### 13.4 PostgreSQL integration tests

```bash
make test-integration
```

The script:

1. starts PostgreSQL;
2. creates a uniquely named isolated test database;
3. applies all 17 migrations;
4. tests auth-session consumption, ledger constraints and rollback, FD atomicity
   and protection, notices, price ingestion, and snapshot immutability/jobs;
5. rolls migration 17 down and up to verify reversibility;
6. drops the temporary database.

It refuses to use a database whose name does not match the isolated test prefix.

### 13.5 Frontend tests and production validation

```bash
make frontend-check
```

This runs ESLint, Node tests, and a full optimized Next.js build. Frontend tests
cover safe navigation, CSV quoting, filename safety, spreadsheet injection
protection, and the multi-currency summary export.

### 13.6 Complete automated validation

```bash
make test-all
```

### 13.7 Full-stack container smoke test

Set a unique JWT secret of at least 32 characters in `.env`, then:

```bash
make app-up
make app-logs
```

Verify:

- `http://localhost:8080/healthz` returns `{"status":"ok"}`;
- `http://localhost:3000` renders;
- registration and login work;
- a portfolio, account, deposit, asset, purchase, and price can be recorded;
- holdings, valuation, allocation, exports, and notices respond;
- restarting API/frontend containers does not lose database data.

Stop without deleting the database volume:

```bash
make app-down
```

## 14. Suggested end-to-end acceptance scenario

1. Register a user with INR base currency and Asia/Kolkata timezone.
2. Create an INR portfolio.
3. Add one brokerage account and one bank account.
4. Create an INR equity asset and an INR mutual-fund asset.
5. Deposit cash into the brokerage account.
6. Buy both assets and confirm holdings and reduced cash.
7. Add a price to only one asset and confirm incomplete valuation plus a missing
   price notice.
8. Add the second price and confirm complete value and allocation.
9. Export the ledger and portfolio summary.
10. Create two daily snapshots around a deposit and price change.
11. Verify performance, risk, and contribution analysis.
12. Create a benchmark with exact boundary observations and verify comparison.
13. Add enough aligned dates to calculate beta.
14. Run concentration, diversification, rebalancing, and health score.
15. Create a goal and a month-end goal snapshot; verify planning and notices.
16. Run SIP and what-if scenarios and confirm nothing is added to portfolio
    holdings or value.
17. Create an FD in the bank account, append a value, then close it; verify the
    ledger, cash proceeds, lifecycle, and notice removal.
18. Reverse and correct an ordinary transaction; verify history remains intact.
19. Attempt to reverse an FD event and verify rejection.
20. Repeat all scheduled commands and confirm idempotency.

## 15. Deployment topology and current readiness

The repository contains production-style non-root API and frontend images and a
local Compose overlay. The current Compose database uses development credentials,
publishes PostgreSQL and API ports, and disables database TLS. It must not be
exposed publicly as-is.

Before public deployment:

- choose a hosting platform and domain;
- use managed PostgreSQL with TLS and unique credentials;
- expose only an HTTPS reverse proxy/frontend publicly;
- store JWT, database, and Upstox secrets in a secret manager;
- configure backups and prove a restore;
- schedule and monitor all four job modes;
- obtain an Upstox Analytics Token and run a live equity ingestion test;
- add CI/CD because no `.github/workflows` currently exists;
- add structured request/application logging, metrics, uptime checks, and alerts;
- perform a staging migration, smoke, restart, backup/restore, and security test.

## 16. Explicit limitations and pending scope

- No broker or bank synchronization.
- No trade execution.
- No AI, prediction, or personalized financial advice.
- No automatic currency conversion or exchange-rate history.
- No inferred prices or benchmark interpolation.
- No automated tax liability calculation despite tax ledger entry support.
- No email, SMS, or push delivery; notices are currently in-app and derived.
- No persisted read/dismiss state for notices.
- No household aggregation, mobile app, or advisor mode.
- No final cloud deployment configuration, CI/CD pipeline, backup policy, or
  production observability stack yet.
- Upstox integration is implemented and unit-tested but still needs a real token
  and live ingestion verification before deployment.

## 17. Key source files

- Project rules: `AGENTS.md`
- Setup and operating commands: `README.md`, `Makefile`
- API composition: `backend/internal/server/server.go`
- Financial formulas: `backend/pkg/finance/`
- Database history: `backend/migrations/`
- Frontend routes: `frontend/src/app/`
- Full-stack containers: `compose.yaml`, `compose.app.yaml`
- Integration harness: `scripts/test-backend-integration.sh`
- Specialized requirements: `docs/fixed-deposits.md`,
  `docs/transaction-csv-import.md`, `docs/notifications.md`, and
  `docs/portfolio-summary-export.md`

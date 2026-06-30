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

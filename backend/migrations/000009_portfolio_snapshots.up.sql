CREATE TABLE portfolio_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE RESTRICT,
    snapshot_date date NOT NULL,
    snapshot_period text NOT NULL CHECK (snapshot_period IN ('daily')),
    total_values jsonb NOT NULL,
    asset_allocations jsonb NOT NULL,
    asset_class_allocations jsonb NOT NULL,
    cash_allocations jsonb NOT NULL,
    missing_prices jsonb NOT NULL,
    is_fully_valued boolean NOT NULL,
    valuation_scope text NOT NULL,
    allocation_scope text NOT NULL,
    valuation_metadata jsonb NOT NULL,
    allocation_metadata jsonb NOT NULL,
    holdings_metadata jsonb NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_portfolio_snapshots_portfolio_date_period
    ON portfolio_snapshots (portfolio_id, snapshot_date, snapshot_period);
CREATE INDEX idx_portfolio_snapshots_portfolio_date
    ON portfolio_snapshots (portfolio_id, snapshot_date DESC);
CREATE INDEX idx_portfolio_snapshots_created_by_user_id
    ON portfolio_snapshots (created_by_user_id);

CREATE OR REPLACE FUNCTION prevent_portfolio_snapshot_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'portfolio snapshots are immutable';
END;
$$;

CREATE TRIGGER trg_portfolio_snapshots_immutable
    BEFORE UPDATE OR DELETE ON portfolio_snapshots
    FOR EACH ROW
    EXECUTE FUNCTION prevent_portfolio_snapshot_mutation();

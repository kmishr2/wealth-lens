CREATE TABLE weekly_performance_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE RESTRICT,
    week_start_date date NOT NULL,
    week_end_date date NOT NULL,
    currency_returns jsonb NOT NULL,
    performance_scope text NOT NULL,
    pnl_metadata jsonb NOT NULL,
    cagr_metadata jsonb NOT NULL,
    xirr_metadata jsonb NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_weekly_performance_snapshots_week_range CHECK (week_end_date > week_start_date)
);

CREATE UNIQUE INDEX ux_weekly_performance_snapshots_portfolio_week_end
    ON weekly_performance_snapshots (portfolio_id, week_end_date);
CREATE INDEX idx_weekly_performance_snapshots_portfolio_week_end
    ON weekly_performance_snapshots (portfolio_id, week_end_date DESC);
CREATE INDEX idx_weekly_performance_snapshots_created_by_user_id
    ON weekly_performance_snapshots (created_by_user_id);

CREATE OR REPLACE FUNCTION prevent_weekly_performance_snapshot_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'weekly performance snapshots are immutable';
END;
$$;

CREATE TRIGGER trg_weekly_performance_snapshots_immutable
    BEFORE UPDATE OR DELETE ON weekly_performance_snapshots
    FOR EACH ROW
    EXECUTE FUNCTION prevent_weekly_performance_snapshot_mutation();

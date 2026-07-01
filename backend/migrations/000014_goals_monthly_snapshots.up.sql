CREATE TABLE goals (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE RESTRICT,
    name text NOT NULL,
    target_amount numeric(28,10) NOT NULL CHECK (target_amount > 0),
    currency char(3) NOT NULL,
    target_date date NOT NULL,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'archived')),
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT ck_goals_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT ck_goals_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT ux_goals_id_portfolio UNIQUE (id, portfolio_id)
);

CREATE UNIQUE INDEX ux_goals_portfolio_name_active
    ON goals (portfolio_id, lower(name))
    WHERE deleted_at IS NULL;
CREATE INDEX idx_goals_portfolio_status
    ON goals (portfolio_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_goals_created_by_user_id ON goals (created_by_user_id);

CREATE TABLE monthly_goal_snapshots (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id uuid NOT NULL REFERENCES portfolios(id) ON DELETE RESTRICT,
    goal_id uuid NOT NULL,
    snapshot_month_end date NOT NULL,
    current_value numeric(28,10) NOT NULL CHECK (current_value >= 0),
    target_value numeric(28,10) NOT NULL CHECK (target_value > 0),
    currency char(3) NOT NULL,
    progress_percentage numeric(28,10) NOT NULL,
    remaining_amount numeric(28,10) NOT NULL CHECK (remaining_amount >= 0),
    months_remaining integer NOT NULL CHECK (months_remaining >= 0),
    required_monthly_contribution numeric(28,10) NOT NULL CHECK (required_monthly_contribution >= 0),
    is_target_reached boolean NOT NULL,
    goal_progress_metadata jsonb NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_monthly_goal_snapshots_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT ck_monthly_goal_snapshots_month_end CHECK (extract(day FROM snapshot_month_end + 1) = 1),
    CONSTRAINT fk_monthly_goal_snapshots_goal_portfolio
        FOREIGN KEY (goal_id, portfolio_id) REFERENCES goals(id, portfolio_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX ux_monthly_goal_snapshots_goal_month
    ON monthly_goal_snapshots (goal_id, snapshot_month_end);
CREATE INDEX idx_monthly_goal_snapshots_portfolio_month
    ON monthly_goal_snapshots (portfolio_id, snapshot_month_end DESC);
CREATE INDEX idx_monthly_goal_snapshots_created_by_user_id
    ON monthly_goal_snapshots (created_by_user_id);

CREATE OR REPLACE FUNCTION prevent_monthly_goal_snapshot_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'monthly goal snapshots are immutable';
END;
$$;

CREATE TRIGGER trg_monthly_goal_snapshots_immutable
    BEFORE UPDATE OR DELETE ON monthly_goal_snapshots
    FOR EACH ROW
    EXECUTE FUNCTION prevent_monthly_goal_snapshot_mutation();

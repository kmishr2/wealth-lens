CREATE TABLE benchmarks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code text NOT NULL,
    name text NOT NULL,
    currency char(3) NOT NULL,
    source text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_benchmarks_code_not_blank CHECK (length(btrim(code)) > 0),
    CONSTRAINT ck_benchmarks_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT ck_benchmarks_currency_format CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT ck_benchmarks_source_not_blank CHECK (length(btrim(source)) > 0)
);

CREATE UNIQUE INDEX ux_benchmarks_code ON benchmarks (upper(code));
CREATE INDEX idx_benchmarks_created_by_user_id ON benchmarks (created_by_user_id);

CREATE TABLE benchmark_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    benchmark_id uuid NOT NULL REFERENCES benchmarks(id) ON DELETE RESTRICT,
    observation_date date NOT NULL,
    value numeric(28,10) NOT NULL CHECK (value > 0),
    source text NOT NULL,
    note text NOT NULL DEFAULT '',
    created_by_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ck_benchmark_observations_source_not_blank CHECK (length(btrim(source)) > 0)
);

CREATE UNIQUE INDEX ux_benchmark_observations_benchmark_date
    ON benchmark_observations (benchmark_id, observation_date);
CREATE INDEX idx_benchmark_observations_benchmark_date
    ON benchmark_observations (benchmark_id, observation_date DESC);
CREATE INDEX idx_benchmark_observations_created_by_user_id
    ON benchmark_observations (created_by_user_id);

CREATE OR REPLACE FUNCTION prevent_benchmark_observation_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'benchmark observations are immutable';
END;
$$;

CREATE TRIGGER trg_benchmark_observations_immutable
    BEFORE UPDATE OR DELETE ON benchmark_observations
    FOR EACH ROW
    EXECUTE FUNCTION prevent_benchmark_observation_mutation();

DROP TRIGGER IF EXISTS trg_portfolio_snapshots_immutable ON portfolio_snapshots;
DROP FUNCTION IF EXISTS prevent_portfolio_snapshot_mutation();
DROP TABLE IF EXISTS portfolio_snapshots;

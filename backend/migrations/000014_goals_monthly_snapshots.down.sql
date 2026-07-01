DROP TRIGGER IF EXISTS trg_monthly_goal_snapshots_immutable ON monthly_goal_snapshots;
DROP FUNCTION IF EXISTS prevent_monthly_goal_snapshot_mutation();
DROP TABLE IF EXISTS monthly_goal_snapshots;
DROP TABLE IF EXISTS goals;

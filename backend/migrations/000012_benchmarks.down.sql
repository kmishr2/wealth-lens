DROP TRIGGER IF EXISTS trg_benchmark_observations_immutable ON benchmark_observations;
DROP FUNCTION IF EXISTS prevent_benchmark_observation_mutation();
DROP TABLE IF EXISTS benchmark_observations;
DROP TABLE IF EXISTS benchmarks;

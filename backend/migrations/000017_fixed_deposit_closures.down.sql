DROP TABLE IF EXISTS fixed_deposit_closures;
ALTER TABLE fixed_deposits DROP CONSTRAINT IF EXISTS uq_fixed_deposits_scope_currency;

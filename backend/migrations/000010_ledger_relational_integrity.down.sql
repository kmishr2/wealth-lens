ALTER TABLE portfolio_snapshots
    DROP CONSTRAINT fk_portfolio_snapshots_creator_portfolio;

ALTER TABLE portfolio_snapshots
    ADD CONSTRAINT portfolio_snapshots_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT;

ALTER TABLE transactions
    DROP CONSTRAINT ck_transactions_distinct_references,
    DROP CONSTRAINT ck_transactions_reversal_shape,
    DROP CONSTRAINT fk_transactions_correction_target,
    DROP CONSTRAINT fk_transactions_reversal_target,
    DROP CONSTRAINT fk_transactions_creator_portfolio,
    DROP CONSTRAINT fk_transactions_account_portfolio;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_account_id_fkey
        FOREIGN KEY (account_id)
        REFERENCES accounts (id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT transactions_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT transactions_reverses_transaction_id_fkey
        FOREIGN KEY (reverses_transaction_id)
        REFERENCES transactions (id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT transactions_corrects_transaction_id_fkey
        FOREIGN KEY (corrects_transaction_id)
        REFERENCES transactions (id)
        ON DELETE RESTRICT;

ALTER TABLE transactions
    DROP CONSTRAINT uq_transactions_id_portfolio_account,
    DROP CONSTRAINT uq_transactions_id_portfolio;

ALTER TABLE accounts
    DROP CONSTRAINT uq_accounts_id_portfolio;

ALTER TABLE portfolios
    DROP CONSTRAINT uq_portfolios_id_user;

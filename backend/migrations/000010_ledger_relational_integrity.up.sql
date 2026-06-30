ALTER TABLE portfolios
    ADD CONSTRAINT uq_portfolios_id_user UNIQUE (id, user_id);

ALTER TABLE accounts
    ADD CONSTRAINT uq_accounts_id_portfolio UNIQUE (id, portfolio_id);

ALTER TABLE transactions
    ADD CONSTRAINT uq_transactions_id_portfolio UNIQUE (id, portfolio_id),
    ADD CONSTRAINT uq_transactions_id_portfolio_account UNIQUE (id, portfolio_id, account_id);

ALTER TABLE transactions
    DROP CONSTRAINT transactions_account_id_fkey,
    DROP CONSTRAINT transactions_created_by_user_id_fkey,
    DROP CONSTRAINT transactions_reverses_transaction_id_fkey,
    DROP CONSTRAINT transactions_corrects_transaction_id_fkey;

ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_account_portfolio
        FOREIGN KEY (account_id, portfolio_id)
        REFERENCES accounts (id, portfolio_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_transactions_creator_portfolio
        FOREIGN KEY (portfolio_id, created_by_user_id)
        REFERENCES portfolios (id, user_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_transactions_reversal_target
        FOREIGN KEY (reverses_transaction_id, portfolio_id, account_id)
        REFERENCES transactions (id, portfolio_id, account_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_transactions_correction_target
        FOREIGN KEY (corrects_transaction_id, portfolio_id)
        REFERENCES transactions (id, portfolio_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT ck_transactions_reversal_shape CHECK (
        (
            transaction_type = 'reversal'
            AND reverses_transaction_id IS NOT NULL
            AND corrects_transaction_id IS NULL
        )
        OR (
            transaction_type <> 'reversal'
            AND reverses_transaction_id IS NULL
        )
    ),
    ADD CONSTRAINT ck_transactions_distinct_references CHECK (
        (reverses_transaction_id IS NULL OR reverses_transaction_id <> id)
        AND (corrects_transaction_id IS NULL OR corrects_transaction_id <> id)
    );

ALTER TABLE portfolio_snapshots
    DROP CONSTRAINT portfolio_snapshots_created_by_user_id_fkey;

ALTER TABLE portfolio_snapshots
    ADD CONSTRAINT fk_portfolio_snapshots_creator_portfolio
        FOREIGN KEY (portfolio_id, created_by_user_id)
        REFERENCES portfolios (id, user_id)
        ON DELETE RESTRICT;

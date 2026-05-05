CREATE TABLE ledger_entries (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id   UUID        NOT NULL REFERENCES transfers(id),
    account_id    UUID        NOT NULL REFERENCES accounts(id),
    amount        BIGINT      NOT NULL CHECK (amount > 0),
    entry_type    VARCHAR(6)  NOT NULL CHECK (entry_type IN ('debit', 'credit')),
    balance_after BIGINT      NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

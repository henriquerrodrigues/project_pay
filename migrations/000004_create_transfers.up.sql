CREATE TABLE transfers (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    end_to_end_id       TEXT        NOT NULL UNIQUE,
    sender_account_id   UUID        NOT NULL REFERENCES accounts(id),
    receiver_account_id UUID        NOT NULL REFERENCES accounts(id),
    amount              BIGINT      NOT NULL CHECK (amount > 0),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ
);

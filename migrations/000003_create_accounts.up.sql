CREATE TABLE accounts (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id),
    participant_id UUID        NOT NULL REFERENCES participants(id),
    account_number TEXT        NOT NULL,
    branch         TEXT        NOT NULL,
    account_type   VARCHAR(4)  NOT NULL CHECK (account_type IN ('cacc', 'svgs')),
    balance        BIGINT      NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

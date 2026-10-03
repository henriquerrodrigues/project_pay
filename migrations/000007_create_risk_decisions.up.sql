-- trilha de auditoria: cada transferência avaliada gera uma linha, aprovada ou não
CREATE TABLE risk_decisions (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id           UUID        REFERENCES transfers(id),
    sender_account_id     UUID        NOT NULL REFERENCES accounts(id),
    receiver_account_id   UUID        NOT NULL REFERENCES accounts(id),
    amount                BIGINT      NOT NULL CHECK (amount > 0),
    decision              VARCHAR(7)  NOT NULL CHECK (decision IN ('APPROVE', 'REVIEW', 'BLOCK')),
    fraud_score           SMALLINT    NOT NULL CHECK (fraud_score BETWEEN 0 AND 100),
    fraud_level           VARCHAR(8)  NOT NULL CHECK (fraud_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    red_flags             TEXT[]      NOT NULL DEFAULT '{}',
    compliance_violations TEXT[]      NOT NULL DEFAULT '{}',
    explanation           TEXT        NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_risk_decisions_sender_created ON risk_decisions (sender_account_id, created_at);

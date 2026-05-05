CREATE TABLE pix_keys (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key_type   VARCHAR(5)  NOT NULL CHECK (key_type IN ('cpf', 'cnpj', 'email', 'phone', 'evp')),
    key_value  TEXT        NOT NULL UNIQUE,
    account_id UUID        NOT NULL REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

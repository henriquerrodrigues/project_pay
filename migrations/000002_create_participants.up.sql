CREATE TABLE participants (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ispb       CHAR(8)     NOT NULL UNIQUE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

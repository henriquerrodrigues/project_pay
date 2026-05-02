CREATE TABLE "users"(
    "id" UUID NOT NULL,
    "name" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "password_hash" TEXT NOT NULL,
    "tax_id" TEXT NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "users" ADD PRIMARY KEY("id");
CREATE TABLE "participants"(
    "id" UUID NOT NULL,
    "ispb" CHAR(8) NOT NULL,
    "name" TEXT NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "participants" ADD PRIMARY KEY("id");
CREATE TABLE "transfers"(
    "id" UUID NOT NULL,
    "end_to_end_id" TEXT NOT NULL,
    "sender_account_id" UUID NOT NULL,
    "receiver_account_id" UUID NOT NULL,
    "amount" BIGINT NOT NULL,
    "status" VARCHAR(255) CHECK
        (
            "status" IN(
                'pending',
                'processing',
                'completed',
                'failed'
            )
        ) NOT NULL,
        "description" TEXT NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "completed_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "transfers" ADD PRIMARY KEY("id");
CREATE TABLE "accounts"(
    "id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "participant_id" UUID NOT NULL,
    "account_number" TEXT NOT NULL,
    "branch" TEXT NOT NULL,
    "account_type" VARCHAR(255) CHECK
        ("account_type" IN('cacc', 'svgs')) NOT NULL,
        "balance" BIGINT NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "accounts" ADD PRIMARY KEY("id");
CREATE TABLE "ledger_entries"(
    "id" UUID NOT NULL,
    "transfer_id" UUID NOT NULL,
    "account_id" UUID NOT NULL,
    "amount" BIGINT NOT NULL,
    "entry_type" VARCHAR(255) CHECK
        ("entry_type" IN('debit', 'credit')) NOT NULL,
        "balance_after" BIGINT NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "ledger_entries" ADD PRIMARY KEY("id");
CREATE TABLE "pix_keys"(
    "id" UUID NOT NULL,
    "key_type" VARCHAR(255) CHECK
        (
            "key_type" IN('cpf', 'cnpj', 'email', 'phone', 'evp')
        ) NOT NULL,
        "key_value" TEXT NOT NULL,
        "account_id" UUID NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "pix_keys" ADD PRIMARY KEY("id");
ALTER TABLE
    "transfers" ADD CONSTRAINT "transfers_receiver_account_id_foreign" FOREIGN KEY("receiver_account_id") REFERENCES "accounts"("participant_id");
ALTER TABLE
    "accounts" ADD CONSTRAINT "accounts_participant_id_foreign" FOREIGN KEY("participant_id") REFERENCES "participants"("id");
ALTER TABLE
    "transfers" ADD CONSTRAINT "transfers_sender_account_id_foreign" FOREIGN KEY("sender_account_id") REFERENCES "accounts"("user_id");
ALTER TABLE
    "ledger_entries" ADD CONSTRAINT "ledger_entries_account_id_foreign" FOREIGN KEY("account_id") REFERENCES "accounts"("user_id");
ALTER TABLE
    "accounts" ADD CONSTRAINT "accounts_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "pix_keys" ADD CONSTRAINT "pix_keys_account_id_foreign" FOREIGN KEY("account_id") REFERENCES "accounts"("user_id");
ALTER TABLE
    "ledger_entries" ADD CONSTRAINT "ledger_entries_entry_type_foreign" FOREIGN KEY("entry_type") REFERENCES "transfers"("end_to_end_id");
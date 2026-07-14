-- scripts/dev_seed.sql
-- Seed de desenvolvimento. Rode após as migrations:
--   docker exec -i project_pay_db psql -U pay_user -d project_pay < scripts/dev_seed.sql
--
-- Participante fixo usado como default no form "Nova conta" do frontend.
INSERT INTO participants (id, ispb, name)
VALUES ('00000000-0000-0000-0000-000000000001', '12345678', 'Banco Dev')
ON CONFLICT (id) DO NOTHING;

-- Contas novas nascem com balance = 0 e não existe endpoint de depósito.
-- Para testar transferência, credite saldo manualmente (valor em centavos):
--   UPDATE accounts SET balance = balance + 100000 WHERE id = '<ACCOUNT_ID>';

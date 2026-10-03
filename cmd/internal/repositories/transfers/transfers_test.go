package transfers_test

// Testes de integração do repositório de transfers contra um Postgres de verdade.
// Os testes de regra de negócio (dono da conta, mesma conta, conta inexistente)
// ficam em services/transfers com fakes; aqui o foco é o que só o banco garante:
// saldo, rollback da transação, ledger_entries e o lock FOR UPDATE.
//
// Rodam só quando TEST_DATABASE_URL está definida, por exemplo:
//   TEST_DATABASE_URL="postgresql://pay_user:pay_pass@localhost:5433/project_pay?sslmode=disable" go test ./...
// Cada teste cria um schema próprio, aplica as migrations e apaga o schema no fim,
// então não mexe nos dados que já estão no banco.

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"project_pay/cmd/internal/models"
	transfersrepo "project_pay/cmd/internal/repositories/transfers"
)

const migrationsDir = "../../../../migrations"

// newTestDB cria um schema isolado, aplica as migrations nele e devolve um pool
// que já usa esse schema como search_path.
func newTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL não definida; pulando teste de integração")
	}
	ctx := context.Background()

	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("conectando no banco: %v", err)
	}
	t.Cleanup(admin.Close)

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("criando schema: %v", err)
	}
	t.Cleanup(func() {
		admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL inválida: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()

	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatalf("conectando no schema de teste: %v", err)
	}
	t.Cleanup(pool.Close)

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("migrations não encontradas em %s: %v", migrationsDir, err)
	}
	sort.Strings(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("lendo %s: %v", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("aplicando %s: %v", filepath.Base(f), err)
		}
	}
	return pool
}

// createAccount insere um user e uma conta com o saldo informado (em centavos).
func createAccount(t *testing.T, db *pgxpool.Pool, participantID uuid.UUID, balance int64) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.NewString()

	var userID uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, tax_id)
		VALUES ('Teste', $1, 'hash', $2)
		RETURNING id`, suffix+"@teste.com", suffix).Scan(&userID)
	if err != nil {
		t.Fatalf("criando user: %v", err)
	}

	var accountID uuid.UUID
	err = db.QueryRow(ctx, `
		INSERT INTO accounts (user_id, participant_id, account_number, branch, account_type, balance)
		VALUES ($1, $2, '12345', '0001', 'cacc', $3)
		RETURNING id`, userID, participantID, balance).Scan(&accountID)
	if err != nil {
		t.Fatalf("criando conta: %v", err)
	}
	return accountID
}

func createParticipant(t *testing.T, db *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := db.QueryRow(context.Background(), `
		INSERT INTO participants (ispb, name) VALUES ('12345678', 'Banco Teste')
		RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("criando participant: %v", err)
	}
	return id
}

func balanceOf(t *testing.T, db *pgxpool.Pool, accountID uuid.UUID) int64 {
	t.Helper()
	var balance int64
	if err := db.QueryRow(context.Background(),
		"SELECT balance FROM accounts WHERE id = $1", accountID).Scan(&balance); err != nil {
		t.Fatalf("lendo saldo: %v", err)
	}
	return balance
}

func countRows(t *testing.T, db *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("contando %s: %v", table, err)
	}
	return n
}

func TestCreate_OKMovimentaSaldoEGeraLedger(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 1000)
	receiver := createAccount(t, db, participant, 200)
	repo := transfersrepo.New(db)

	transfer, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: sender, ReceiverAccountID: receiver, Amount: 300, Description: "almoço",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if transfer.Status != "completed" {
		t.Errorf("status = %q, esperava completed", transfer.Status)
	}
	if !strings.HasPrefix(transfer.EndToendID, "E12345678") {
		t.Errorf("end_to_end_id = %q, esperava prefixo E + ISPB", transfer.EndToendID)
	}

	if got := balanceOf(t, db, sender); got != 700 {
		t.Errorf("saldo origem = %d, esperava 700", got)
	}
	if got := balanceOf(t, db, receiver); got != 500 {
		t.Errorf("saldo destino = %d, esperava 500", got)
	}

	rows, err := db.Query(context.Background(), `
		SELECT account_id, entry_type, amount, balance_after
		FROM ledger_entries WHERE transfer_id = $1`, transfer.ID)
	if err != nil {
		t.Fatalf("lendo ledger: %v", err)
	}
	defer rows.Close()

	type entry struct {
		account      uuid.UUID
		amount       int64
		balanceAfter int64
	}
	entries := map[string]entry{}
	for rows.Next() {
		var e entry
		var kind string
		if err := rows.Scan(&e.account, &kind, &e.amount, &e.balanceAfter); err != nil {
			t.Fatalf("scan ledger: %v", err)
		}
		entries[kind] = e
	}
	if len(entries) != 2 {
		t.Fatalf("esperava 1 debit e 1 credit, veio %v", entries)
	}
	if e := entries["debit"]; e.account != sender || e.amount != 300 || e.balanceAfter != 700 {
		t.Errorf("debit errado: %+v", e)
	}
	if e := entries["credit"]; e.account != receiver || e.amount != 300 || e.balanceAfter != 500 {
		t.Errorf("credit errado: %+v", e)
	}
}

func TestCreate_SaldoInsuficienteNaoMexeEmNada(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 50)
	receiver := createAccount(t, db, participant, 0)
	repo := transfersrepo.New(db)

	_, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: sender, ReceiverAccountID: receiver, Amount: 100,
	})
	if !errors.Is(err, transfersrepo.ErrInsufficientFunds) {
		t.Fatalf("esperava ErrInsufficientFunds, veio %v", err)
	}

	if got := balanceOf(t, db, sender); got != 50 {
		t.Errorf("saldo origem = %d, esperava 50", got)
	}
	if got := balanceOf(t, db, receiver); got != 0 {
		t.Errorf("saldo destino = %d, esperava 0", got)
	}
	if n := countRows(t, db, "transfers"); n != 0 {
		t.Errorf("transfers = %d, esperava 0", n)
	}
	if n := countRows(t, db, "ledger_entries"); n != 0 {
		t.Errorf("ledger_entries = %d, esperava 0", n)
	}
}

func TestCreate_SaldoExatoZeraConta(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 100)
	receiver := createAccount(t, db, participant, 0)
	repo := transfersrepo.New(db)

	if _, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: sender, ReceiverAccountID: receiver, Amount: 100,
	}); err != nil {
		t.Fatalf("transferir o saldo inteiro deveria funcionar: %v", err)
	}
	if got := balanceOf(t, db, sender); got != 0 {
		t.Errorf("saldo origem = %d, esperava 0", got)
	}
}

func TestCreate_OrigemInexistente(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	receiver := createAccount(t, db, participant, 0)
	repo := transfersrepo.New(db)

	_, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: uuid.New(), ReceiverAccountID: receiver, Amount: 100,
	})
	if err == nil {
		t.Fatal("esperava erro para conta de origem inexistente")
	}
	if n := countRows(t, db, "transfers"); n != 0 {
		t.Errorf("transfers = %d, esperava 0", n)
	}
}

// O débito da origem acontece antes de buscar o destino; se o destino não
// existe, a transação inteira precisa voltar e a origem continua com o saldo.
func TestCreate_DestinoInexistenteDesfazDebito(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 1000)
	repo := transfersrepo.New(db)

	_, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: sender, ReceiverAccountID: uuid.New(), Amount: 100,
	})
	if err == nil {
		t.Fatal("esperava erro para conta de destino inexistente")
	}
	if got := balanceOf(t, db, sender); got != 1000 {
		t.Errorf("saldo origem = %d, esperava 1000 (débito deveria ter sido desfeito)", got)
	}
	if n := countRows(t, db, "ledger_entries"); n != 0 {
		t.Errorf("ledger_entries = %d, esperava 0", n)
	}
}

// Várias transferências simultâneas da mesma conta não podem gastar mais do
// que o saldo: o SELECT ... FOR UPDATE serializa a leitura do saldo.
func TestCreate_ConcorrenciaNaoDeixaSaldoNegativo(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 100)
	receiver := createAccount(t, db, participant, 0)
	repo := transfersrepo.New(db)

	const workers = 10
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := repo.Create(context.Background(), models.CreateTransferRequest{
				SenderAccountID: sender, ReceiverAccountID: receiver, Amount: 30,
				Description: fmt.Sprintf("tentativa %d", i),
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)

	ok, insufficient := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, transfersrepo.ErrInsufficientFunds):
			insufficient++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 3 || insufficient != workers-3 {
		t.Errorf("sucessos = %d, saldo insuficiente = %d; esperava 3 e %d", ok, insufficient, workers-3)
	}
	if got := balanceOf(t, db, sender); got != 10 {
		t.Errorf("saldo origem = %d, esperava 10", got)
	}
	if got := balanceOf(t, db, receiver); got != 90 {
		t.Errorf("saldo destino = %d, esperava 90", got)
	}
}

func TestGetByID(t *testing.T) {
	db := newTestDB(t)
	participant := createParticipant(t, db)
	sender := createAccount(t, db, participant, 1000)
	receiver := createAccount(t, db, participant, 0)
	repo := transfersrepo.New(db)

	created, err := repo.Create(context.Background(), models.CreateTransferRequest{
		SenderAccountID: sender, ReceiverAccountID: receiver, Amount: 100,
	})
	if err != nil {
		t.Fatalf("criando transfer: %v", err)
	}

	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.ID != created.ID || got.Amount != 100 || got.EndToendID != created.EndToendID {
		t.Errorf("transfer lida diferente da criada: %+v", got)
	}

	missing, err := repo.GetByID(context.Background(), uuid.New())
	if err != nil || missing != nil {
		t.Errorf("id inexistente deveria voltar (nil, nil), veio (%v, %v)", missing, err)
	}
}

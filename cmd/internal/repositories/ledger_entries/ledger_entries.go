package ledgerentries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"project_pay/cmd/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.LedgerEntry, error) {
	rows, err := r.db.Query(ctx, sqlGetByAccountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.LedgerEntry{}
	for rows.Next() {
		var e models.LedgerEntry
		if err := rows.Scan(
			&e.ID,
			&e.TransferID,
			&e.AccountID,
			&e.Amount,
			&e.EntryType,
			&e.BalanceAfter,
			&e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

const sqlGetByAccountID = `
	SELECT id, transfer_id, account_id, amount, entry_type, balance_after, created_at
	FROM ledger_entries
	WHERE account_id = $1
	ORDER BY created_at DESC`

package repository

import (
	"context"
	"financial-api/internal/entity"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)

	GetUserByIDTx(ctx context.Context, tx pgx.Tx, userID int) (*entity.User, error)
	UpdateUserBalanceTx(ctx context.Context, tx pgx.Tx, userID int, newBalance float64) error
	CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID int, amount float64, ttype string) error

	GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error)
}

type PgRepository struct {
	pool *pgxpool.Pool
}

func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *PgRepository) GetUserByIDTx(ctx context.Context, tx pgx.Tx, userID int) (*entity.User, error) {
	var u entity.User
	err := tx.QueryRow(ctx, "SELECT id, balance FROM users WHERE id = $1", userID).
		Scan(&u.ID, &u.Balance)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PgRepository) UpdateUserBalanceTx(ctx context.Context, tx pgx.Tx, userID int, newBalance float64) error {
	_, err := tx.Exec(ctx, "UPDATE users SET balance = $1 WHERE id = $2", newBalance, userID)
	return err
}

func (r *PgRepository) CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID int, amount float64, ttype string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO transactions (user_id, amount, type)
		 VALUES ($1, $2, $3)`,
		userID, amount, ttype)
	return err
}

func (r *PgRepository) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, type, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entity.Transaction
	for rows.Next() {
		var tr entity.Transaction
		if err := rows.Scan(&tr.ID, &tr.UserID, &tr.Amount, &tr.Type, &tr.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, tr)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

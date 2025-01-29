// repository/repository.go
package repository

import (
	"context"
	"errors" // Импортируем пакет errors

	"financial-api/internal/entity"

	"github.com/jackc/pgx/v4" // Импортируем пакет pgx
	"github.com/jackc/pgx/v4/pgxpool"
)

type Repository interface {
	BeginTx(ctx context.Context) (Transaction, error)
	GetUserByIDTx(ctx context.Context, tx Transaction, userID int) (*entity.User, error)
	UpdateUserBalanceTx(ctx context.Context, tx Transaction, userID int, newBalance float64) error
	CreateTransactionTx(ctx context.Context, tx Transaction, userID int, amount float64, ttype string) error
	GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error)
}

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Реализация репозитория на основе pgxpool
type PgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PgRepository { // Экспорт функции
	return &PgRepository{pool: pool}
}

func (r *PgRepository) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTransaction{tx: tx}, nil
}

func (r *PgRepository) GetUserByIDTx(ctx context.Context, tx Transaction, userID int) (*entity.User, error) {
	pgxTx, ok := tx.(*pgxTransaction)
	if !ok {
		return nil, errors.New("invalid transaction type") // Используем errors.New
	}

	var user entity.User
	query := "SELECT id, balance FROM users WHERE id=$1"
	err := pgxTx.tx.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Balance)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgRepository) UpdateUserBalanceTx(ctx context.Context, tx Transaction, userID int, newBalance float64) error {
	pgxTx, ok := tx.(*pgxTransaction)
	if !ok {
		return errors.New("invalid transaction type") // Используем errors.New
	}

	query := "UPDATE users SET balance=$1 WHERE id=$2"
	_, err := pgxTx.tx.Exec(ctx, query, newBalance, userID)
	return err
}

func (r *PgRepository) CreateTransactionTx(ctx context.Context, tx Transaction, userID int, amount float64, ttype string) error {
	pgxTx, ok := tx.(*pgxTransaction)
	if !ok {
		return errors.New("invalid transaction type") // Используем errors.New
	}

	query := "INSERT INTO transactions (user_id, amount, type) VALUES ($1, $2, $3)"
	_, err := pgxTx.tx.Exec(ctx, query, userID, amount, ttype)
	return err
}

func (r *PgRepository) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, user_id, amount, type, created_at 
        FROM transactions 
        WHERE user_id=$1 
        ORDER BY created_at DESC 
        LIMIT 10`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		var tr entity.Transaction
		if err := rows.Scan(&tr.ID, &tr.UserID, &tr.Amount, &tr.Type, &tr.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tr)
	}
	return transactions, nil
}

// Реализация транзакции на основе pgx
type pgxTransaction struct {
	tx pgx.Tx // Используем тип pgx.Tx
}

func (t *pgxTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

package service

import (
	"context"
	"financial-api/internal/entity"
	"financial-api/internal/repository"
	"fmt"
)

type Service struct {
	repo repository.Repository // <-- Используем интерфейс, а не *Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TopUpBalance(ctx context.Context, userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	tx, err := s.repo.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	user, err := s.repo.GetUserByIDTx(ctx, tx, userID)
	if err != nil {
		return err
	}

	newBalance := user.Balance + amount
	err = s.repo.UpdateUserBalanceTx(ctx, tx, userID, newBalance)
	if err != nil {
		return err
	}

	err = s.repo.CreateTransactionTx(ctx, tx, userID, amount, "topup")
	if err != nil {
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return commitErr
	}
	return nil
}

func (s *Service) TransferMoney(ctx context.Context, fromUserID, toUserID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to the same user")
	}

	tx, err := s.repo.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	fromUser, err := s.repo.GetUserByIDTx(ctx, tx, fromUserID)
	if err != nil {
		return err
	}
	toUser, err := s.repo.GetUserByIDTx(ctx, tx, toUserID)
	if err != nil {
		return err
	}

	if fromUser.Balance < amount {
		return fmt.Errorf("insufficient funds for user %d", fromUserID)
	}

	fromNew := fromUser.Balance - amount
	toNew := toUser.Balance + amount

	err = s.repo.UpdateUserBalanceTx(ctx, tx, fromUserID, fromNew)
	if err != nil {
		return err
	}
	err = s.repo.UpdateUserBalanceTx(ctx, tx, toUserID, toNew)
	if err != nil {
		return err
	}

	// Записываем 2 транзакции (минус и плюс):
	err = s.repo.CreateTransactionTx(ctx, tx, fromUserID, -amount, "transfer")
	if err != nil {
		return err
	}
	err = s.repo.CreateTransactionTx(ctx, tx, toUserID, amount, "transfer")
	if err != nil {
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return commitErr
	}
	return nil
}

func (s *Service) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	return s.repo.GetLastTransactions(ctx, userID)
}

// service/service.go
package service

import (
	"context"
	"fmt"

	"financial-api/internal/entity"
	"financial-api/internal/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TopUpBalance(ctx context.Context, userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	user, err := s.repo.GetUserByIDTx(ctx, tx, userID)
	if err != nil {
		return err
	}

	newBalance := user.Balance + amount
	if err := s.repo.UpdateUserBalanceTx(ctx, tx, userID, newBalance); err != nil {
		return err
	}

	if err := s.repo.CreateTransactionTx(ctx, tx, userID, amount, "topup"); err != nil {
		return err
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

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
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

	if err := s.repo.UpdateUserBalanceTx(ctx, tx, fromUserID, fromUser.Balance-amount); err != nil {
		return err
	}

	if err := s.repo.UpdateUserBalanceTx(ctx, tx, toUserID, toUser.Balance+amount); err != nil {
		return err
	}

	if err := s.repo.CreateTransactionTx(ctx, tx, fromUserID, -amount, "transfer"); err != nil {
		return err
	}

	if err := s.repo.CreateTransactionTx(ctx, tx, toUserID, amount, "transfer"); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	return s.repo.GetLastTransactions(ctx, userID)
}

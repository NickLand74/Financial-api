package service

import (
	"context"
	"fmt"

	"financial-api/internal/entity"
	"financial-api/internal/repository"
)

// Service — слой бизнес-логики
type Service struct {
	repo repository.Repository
}

// NewService — конструктор Service
func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

// TopUpBalance — пополнение баланса пользователя
func (s *Service) TopUpBalance(ctx context.Context, userID int, amount float64) (err error) {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// Начинаем транзакцию через repo.BeginTx
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	// Если будет ошибка в середине метода, откатываемся
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Вызываем репозиторий с тем же tx (не nil!)
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

	// Всё успешно — делаем commit
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return commitErr
	}

	return nil
}

// TransferMoney — перевод денег от fromUser к toUser
func (s *Service) TransferMoney(ctx context.Context, fromUserID, toUserID int, amount float64) (err error) {
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

	// Обновляем баланс обоих пользователей
	err = s.repo.UpdateUserBalanceTx(ctx, tx, fromUserID, fromUser.Balance-amount)
	if err != nil {
		return err
	}
	err = s.repo.UpdateUserBalanceTx(ctx, tx, toUserID, toUser.Balance+amount)
	if err != nil {
		return err
	}

	// Создаём две записи транзакций (минус у fromUser, плюс у toUser)
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

// GetLastTransactions — возвращает последние 10 транзакций пользователя
func (s *Service) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	return s.repo.GetLastTransactions(ctx, userID)
}

package service_test

import (
	"context"
	"errors"
	"financial-api/internal/entity"
	"financial-api/internal/service"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo — наша реализация interface repository.Repository для тестов.
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	// Возвращаем мок-транзакцию и ошибку (если нужно)
	tx, _ := args.Get(0).(pgx.Tx)
	err, _ := args.Get(1).(error)
	return tx, err
}

func (m *MockRepo) GetUserByIDTx(ctx context.Context, tx pgx.Tx, userID int) (*entity.User, error) {
	args := m.Called(ctx, tx, userID)
	user, _ := args.Get(0).(*entity.User)
	err, _ := args.Get(1).(error)
	return user, err
}

func (m *MockRepo) UpdateUserBalanceTx(ctx context.Context, tx pgx.Tx, userID int, newBalance float64) error {
	args := m.Called(ctx, tx, userID, newBalance)
	return args.Error(0)
}

func (m *MockRepo) CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID int, amount float64, ttype string) error {
	args := m.Called(ctx, tx, userID, amount, ttype)
	return args.Error(0)
}

func (m *MockRepo) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	args := m.Called(ctx, userID)
	trans, _ := args.Get(0).([]entity.Transaction)
	err, _ := args.Get(1).(error)
	return trans, err
}

// MockTx — Фейковая структура для транзакции, которая реализует методы, необходимые для тестирования
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// ===== TESTS =====

func TestTopUpBalance_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	mockTx := new(MockTx)

	// Сценарий:
	// 1) BeginTx -> вернём (mockTx, nil)
	// 2) GetUserByIDTx -> вернём user с балансом 100
	// 3) UpdateUserBalanceTx -> ok
	// 4) CreateTransactionTx -> ok
	// 5) tx.Commit при успешном окончании

	// Настраиваем ожидания:
	mockRepo.On("BeginTx", ctx).Return(mockTx, nil)

	mockRepo.On("GetUserByIDTx", ctx, mockTx, 1).
		Return(&entity.User{ID: 1, Balance: 100.0}, nil)

	mockRepo.On("UpdateUserBalanceTx", ctx, mockTx, 1, 150.0).
		Return(nil)

	mockRepo.On("CreateTransactionTx", ctx, mockTx, 1, 50.0, "topup").
		Return(nil)

	// tx.Commit
	mockTx.On("Commit", ctx).Return(nil)

	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestTopUpBalance_Error(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	// При ошибке подключения, BeginTx должен вернуть ошибку
	mockRepo.On("BeginTx", ctx).Return(nil, errors.New("DB connection error"))

	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB connection error")

	mockRepo.AssertExpectations(t)
}

func TestTransferMoney_Insufficient(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	mockTx := new(MockTx)

	mockRepo.On("BeginTx", ctx).Return(mockTx, nil)

	mockRepo.On("GetUserByIDTx", ctx, mockTx, 1).
		Return(&entity.User{ID: 1, Balance: 20}, nil)
	mockRepo.On("GetUserByIDTx", ctx, mockTx, 2).
		Return(&entity.User{ID: 2, Balance: 100}, nil)

	// Проверка: недостаточно средств, откатим транзакцию
	mockTx.On("Rollback", ctx).Return(nil)

	err := svc.TransferMoney(ctx, 1, 2, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

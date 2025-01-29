// service/service_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"financial-api/internal/entity"
	"financial-api/internal/repository"
	"financial-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository реализует интерфейс repository.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) BeginTx(ctx context.Context) (repository.Transaction, error) {
	args := m.Called(ctx)
	tx, _ := args.Get(0).(repository.Transaction)
	err, _ := args.Get(1).(error)
	return tx, err
}

func (m *MockRepository) GetUserByIDTx(ctx context.Context, tx repository.Transaction, userID int) (*entity.User, error) {
	args := m.Called(ctx, tx, userID)
	u, _ := args.Get(0).(*entity.User)
	err, _ := args.Get(1).(error)
	return u, err
}

func (m *MockRepository) UpdateUserBalanceTx(ctx context.Context, tx repository.Transaction, userID int, newBalance float64) error {
	args := m.Called(ctx, tx, userID, newBalance) // Используем float64
	return args.Error(0)
}

func (m *MockRepository) CreateTransactionTx(ctx context.Context, tx repository.Transaction, userID int, amount float64, ttype string) error {
	args := m.Called(ctx, tx, userID, amount, ttype)
	return args.Error(0)
}

func (m *MockRepository) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	args := m.Called(ctx, userID)
	trs, _ := args.Get(0).([]entity.Transaction)
	err, _ := args.Get(1).(error)
	return trs, err
}

// MockTransaction реализует интерфейс repository.Transaction
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransaction) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// =========== TESTS =============

func TestTopUpBalance_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := service.NewService(mockRepo)
	ctx := context.Background()
	mockTx := new(MockTransaction)

	// Правила моков
	mockRepo.On("BeginTx", ctx).Return(mockTx, nil).Once()
	mockRepo.On("GetUserByIDTx", ctx, mockTx, 1).
		Return(&entity.User{ID: 1, Balance: 100}, nil).
		Once()
	mockRepo.On("UpdateUserBalanceTx", ctx, mockTx, 1, 150.0).
		Return(nil).
		Once()
	mockRepo.On("CreateTransactionTx", ctx, mockTx, 1, 50.0, "topup").
		Return(nil).
		Once()
	mockTx.On("Commit", ctx).Return(nil).Once()
	mockTx.On("Rollback", ctx).Return(nil).Maybe() // Добавляем Maybe(), чтобы Rollback мог быть опциональным

	t.Logf("ExpectedCalls before = %#v", mockRepo.ExpectedCalls)
	t.Log("=== Starting TestTopUpBalance_Success ===")

	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestTransferMoney_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := service.NewService(mockRepo)
	ctx := context.Background()
	mockTx := new(MockTransaction)

	fromUserID := 1
	toUserID := 2
	amount := 50.0

	fromUser := &entity.User{ID: fromUserID, Balance: 150}
	toUser := &entity.User{ID: toUserID, Balance: 100}

	// Правила моков
	mockRepo.On("BeginTx", ctx).Return(mockTx, nil).Once()
	mockRepo.On("GetUserByIDTx", ctx, mockTx, fromUserID).
		Return(fromUser, nil).
		Once()
	mockRepo.On("GetUserByIDTx", ctx, mockTx, toUserID).
		Return(toUser, nil).
		Once()
	mockRepo.On("UpdateUserBalanceTx", ctx, mockTx, fromUserID, 100.0).
		Return(nil).
		Once()
	mockRepo.On("UpdateUserBalanceTx", ctx, mockTx, toUserID, 150.0).
		Return(nil).
		Once()
	mockRepo.On("CreateTransactionTx", ctx, mockTx, fromUserID, -amount, "transfer").
		Return(nil).
		Once()
	mockRepo.On("CreateTransactionTx", ctx, mockTx, toUserID, amount, "transfer").
		Return(nil).
		Once()
	mockTx.On("Commit", ctx).Return(nil).Once()
	mockTx.On("Rollback", ctx).Return(nil).Maybe() // Добавляем Maybe(), чтобы Rollback мог быть опциональным

	t.Logf("ExpectedCalls before = %#v", mockRepo.ExpectedCalls)
	t.Log("=== Starting TestTransferMoney_Success ===")

	err := svc.TransferMoney(ctx, fromUserID, toUserID, amount)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestGetLastTransactions_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := service.NewService(mockRepo)
	ctx := context.Background()

	userID := 1
	expectedTransactions := []entity.Transaction{
		{ID: 1, UserID: userID, Amount: 100, Type: "topup", CreatedAt: time.Now()},
		{ID: 2, UserID: userID, Amount: 50, Type: "transfer", CreatedAt: time.Now()},
	}

	// Правила моков
	mockRepo.On("GetLastTransactions", ctx, userID).
		Return(expectedTransactions, nil).
		Once()

	t.Logf("ExpectedCalls before = %#v", mockRepo.ExpectedCalls)
	t.Log("=== Starting TestGetLastTransactions_Success ===")

	transactions, err := svc.GetLastTransactions(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	mockRepo.AssertExpectations(t)
}

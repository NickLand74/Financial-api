package service_test

import (
	"context"
	"financial-api/internal/entity"
	"financial-api/internal/service"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// MockRepo имплементирует interface repository.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Pool() *pgxpool.Pool {
	// Вернём nil (так как мы не будем реально начинать транзакцию в тесте)
	return nil
}

func (m *MockRepo) GetUserByIDTx(ctx context.Context, tx pgx.Tx, userID int) (*entity.User, error) {
	args := m.Called(ctx, tx, userID)
	// раскладываем
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
	res, _ := args.Get(0).([]entity.Transaction)
	err, _ := args.Get(1).(error)
	return res, err
}

func TestTopUpBalance_Error(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()

	mockRepo.On("Pool").Return(nil)
	mockRepo.On("GetUserByIDTx", ctx, mock.Anything, 1).
		Return(nil, fmt.Errorf("user not found"))

	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")

	mockRepo.AssertExpectations(t)
}

func TestTransferMoney(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()

	mockRepo.On("Pool").Return(nil)

	mockRepo.On("GetUserByIDTx", ctx, mock.Anything, 1).
		Return(&entity.User{ID: 1, Balance: 200}, nil).
		Once()
	mockRepo.On("GetUserByIDTx", ctx, mock.Anything, 2).
		Return(&entity.User{ID: 2, Balance: 100}, nil).
		Once()

	// После списания у 1 баланс станет 150
	mockRepo.On("UpdateUserBalanceTx", ctx, mock.Anything, 1, 150.0).
		Return(nil).
		Once()
	// У 2 станет 150
	mockRepo.On("UpdateUserBalanceTx", ctx, mock.Anything, 2, 150.0).
		Return(nil).
		Once()

	// Две транзакции: -50 для 1, +50 для 2
	mockRepo.On("CreateTransactionTx", ctx, mock.Anything, 1, -50.0, "transfer").
		Return(nil).
		Once()
	mockRepo.On("CreateTransactionTx", ctx, mock.Anything, 2, 50.0, "transfer").
		Return(nil).
		Once()

	err := svc.TransferMoney(ctx, 1, 2, 50.0)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestTransferMoney_Insufficient(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Pool").Return(nil)

	mockRepo.On("GetUserByIDTx", ctx, mock.Anything, 1).
		Return(&entity.User{ID: 1, Balance: 20}, nil)
	mockRepo.On("GetUserByIDTx", ctx, mock.Anything, 2).
		Return(&entity.User{ID: 2, Balance: 100}, nil)

	// Не дойдём до Update, т.к. должно вернуться ошибка
	err := svc.TransferMoney(ctx, 1, 2, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
}

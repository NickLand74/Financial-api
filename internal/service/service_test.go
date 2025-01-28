package service_test

import (
	"context"
	"errors"
	"testing"

	"financial-api/internal/entity"
	"financial-api/internal/service"

	// Если вы НЕ используете "repository" пакет прямо в тестах — удалите импорт:
	// "financial-api/internal/repository"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// === MockRepo ===

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
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
	trs, _ := args.Get(0).([]entity.Transaction)
	err, _ := args.Get(1).(error)
	return trs, err
}

// === MockTx ===

type MockTx struct {
	mock.Mock
}

// Если сервис вызывает tx.Commit
func (mt *MockTx) Commit(ctx context.Context) error {
	args := mt.Called(ctx)
	return args.Error(0)
}

// Если сервис вызывает tx.Rollback
func (mt *MockTx) Rollback(ctx context.Context) error {
	args := mt.Called(ctx)
	return args.Error(0)
}

// Если репозиторий вызывает tx.Exec
// Сейчас pgx.Tx.Exec(...) обычно возвращает (pgconn.CommandTag, error)
func (mt *MockTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	// вернём фиктивный CommandTag
	return pgconn.CommandTag("MOCK"), nil
}

// Если репозиторий или код вызывает Prepare (в новой pgx: Prepare(ctx, name, sql) (*pgconn.StatementDescription, error))
func (mt *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("not implemented")
}

// Ниже методы-заглушки, чтобы полностью удовлетворять интерфейсу pgx.Tx:

func (mt *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}

func (mt *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

func (mt *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, errors.New("not implemented")
}

func (mt *MockTx) Conn() *pgx.Conn {
	return nil
}

// === Тесты ===

func TestTopUpBalance_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	mockTx := new(MockTx)

	mockRepo.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockRepo.On("GetUserByIDTx", mock.Anything, mock.Anything, 1).
		Return(&entity.User{ID: 1, Balance: 100}, nil)
	mockRepo.On("UpdateUserBalanceTx", mock.Anything, mock.Anything, 1, 150.0).
		Return(nil)
	mockRepo.On("CreateTransactionTx", mock.Anything, mock.Anything, 1, 50.0, "topup").
		Return(nil)
	mockTx.On("Commit", mock.Anything).Return(nil)

	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestTopUpBalance_BeginTxError(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	mockRepo.On("BeginTx", mock.Anything).Return(nil, errors.New("DB init error"))

	err := svc.TopUpBalance(context.Background(), 1, 100.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB init error")

	mockRepo.AssertExpectations(t)
}

func TestTransferMoney_Insufficient(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	mockTx := new(MockTx)
	mockRepo.On("BeginTx", mock.Anything).Return(mockTx, nil)

	mockRepo.On("GetUserByIDTx", mock.Anything, mock.Anything, 1).
		Return(&entity.User{ID: 1, Balance: 20}, nil)
	mockRepo.On("GetUserByIDTx", mock.Anything, mock.Anything, 2).
		Return(&entity.User{ID: 2, Balance: 100}, nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	err := svc.TransferMoney(context.Background(), 1, 2, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

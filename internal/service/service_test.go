package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"financial-api/internal/entity"
	"financial-api/internal/service"

	"github.com/jackc/pgconn" // для pgconn.CommandTag
	"github.com/jackc/pgx/v4" // для pgx.Tx
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo имитирует interface repository.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	fmt.Println("MOCK BeginTx called with:", ctx)
	args := m.Called(ctx)
	tx, _ := args.Get(0).(pgx.Tx)
	err, _ := args.Get(1).(error)
	return tx, err
}

func (m *MockRepo) GetUserByIDTx(ctx context.Context, tx pgx.Tx, userID int) (*entity.User, error) {
	fmt.Printf("MOCK GetUserByIDTx called: ctx=%v, tx=%v, userID=%v\n", ctx, tx, userID)
	args := m.Called(ctx, tx, userID)
	u, _ := args.Get(0).(*entity.User)
	err, _ := args.Get(1).(error)
	return u, err
}

func (m *MockRepo) UpdateUserBalanceTx(ctx context.Context, tx pgx.Tx, userID int, newBalance float64) error {
	fmt.Printf("MOCK UpdateUserBalanceTx called: userID=%v, newBalance=%.2f\n", userID, newBalance)
	args := m.Called(ctx, tx, userID, newBalance)
	return args.Error(0)
}

func (m *MockRepo) CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID int, amount float64, ttype string) error {
	fmt.Printf("MOCK CreateTransactionTx called: userID=%v, amount=%.2f, type=%s\n", userID, amount, ttype)
	args := m.Called(ctx, tx, userID, amount, ttype)
	return args.Error(0)
}

func (m *MockRepo) GetLastTransactions(ctx context.Context, userID int) ([]entity.Transaction, error) {
	fmt.Printf("MOCK GetLastTransactions called: userID=%v\n", userID)
	args := m.Called(ctx, userID)
	trs, _ := args.Get(0).([]entity.Transaction)
	err, _ := args.Get(1).(error)
	return trs, err
}

// MockTx имитирует pgx.Tx
type MockTx struct {
	mock.Mock
}

func (mt *MockTx) Commit(ctx context.Context) error {
	fmt.Println("MOCK tx.Commit called")
	args := mt.Called(ctx)
	return args.Error(0)
}
func (mt *MockTx) Rollback(ctx context.Context) error {
	fmt.Println("MOCK tx.Rollback called")
	args := mt.Called(ctx)
	return args.Error(0)
}
func (mt *MockTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	fmt.Println("MOCK tx.Exec called:", sql, args)
	return pgconn.CommandTag("MOCK"), nil
}
func (mt *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	fmt.Println("MOCK tx.Prepare called:", name, sql)
	return nil, errors.New("not implemented")
}
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

// ====================== TESTS =============================

func TestTopUpBalance_Success(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	ctx := context.Background()
	mockTx := new(MockTx)

	// Разрешаем все вызовы (Maybe), с любыми аргументами (mock.Anything)
	mockRepo.On("BeginTx", mock.Anything).
		Maybe().
		Return(mockTx, nil)

	mockRepo.On("GetUserByIDTx", mock.Anything, mock.Anything, mock.Anything).
		Maybe().
		Return(&entity.User{ID: 1, Balance: 100.0}, nil)

	mockRepo.On("UpdateUserBalanceTx", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().
		Return(nil)

	mockRepo.On("CreateTransactionTx", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Maybe().
		Return(nil)

	// Коммит
	mockTx.On("Commit", mock.Anything).Maybe().Return(nil)
	// Роллбек
	mockTx.On("Rollback", mock.Anything).Maybe().Return(nil)

	t.Log("=== Starting TestTopUpBalance_Success ===")
	err := svc.TopUpBalance(ctx, 1, 50.0)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestTopUpBalance_BeginTxError(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	// Если BeginTx вернёт ошибку
	mockRepo.On("BeginTx", mock.Anything).
		Return(nil, errors.New("DB init error"))

	err := svc.TopUpBalance(context.Background(), 1, 100.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB init error")

	mockRepo.AssertExpectations(t)
}

func TestTransferMoney_Insufficient(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := service.NewService(mockRepo)

	mockTx := new(MockTx)

	mockRepo.On("BeginTx", mock.Anything).
		Maybe().
		Return(mockTx, nil)

	// Допустим, для обоих GetUserByIDTx вызовов возвращаем баланс 20 — этого
	// хватит, чтобы сервис понял "insufficient"
	mockRepo.On("GetUserByIDTx", mock.Anything, mock.Anything, mock.Anything).
		Maybe().
		Return(&entity.User{ID: 1, Balance: 20}, nil)

	mockTx.On("Rollback", mock.Anything).
		Maybe().
		Return(nil)

	t.Log("=== Starting TestTransferMoney_Insufficient ===")
	err := svc.TransferMoney(context.Background(), 1, 2, 50.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

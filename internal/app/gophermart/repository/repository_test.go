package repository

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Repository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	repo := New(db)
	return db, mock, repo
}

func TestCreateUser_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE username = \\$1\\)").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec("INSERT INTO users").
		WithArgs("testuser", "hashedpassword").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateUser("testuser", "hashedpassword")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet()) // Ensure all expectations were met
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE username = \\$1\\)").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err := repo.CreateUser("testuser", "hashedpassword")
	assert.ErrorIs(t, err, ErrUserExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPasswordHashByUsername_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT password_hash FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow("hashedpassword"))

	hash, err := repo.GetPasswordHashByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "hashedpassword", hash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPasswordHashByUsername_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT password_hash FROM users WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnError(ErrUserNotFound)

	hash, err := repo.GetPasswordHashByUsername("testuser")
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Empty(t, hash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_NewOrder(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM orders WHERE number = \\$1\\)").
		WithArgs("12345").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectExec("INSERT INTO orders").
		WithArgs("12345", "testuser", "NEW", 0.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	newOrder := models.Order{
		Number:   "12345",
		Username: "testuser",
		Status:   "NEW",
		Accrual:  0.0,
	}

	order, exists, err := repo.CreateOrder(ctx, newOrder)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Equal(t, "12345", order.Number)
	assert.Equal(t, "testuser", order.Username)
	assert.Equal(t, models.StatusNew, order.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOrder_AlreadyExists(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM orders WHERE number = \\$1\\)").
		WithArgs("12345").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery("SELECT number, username, status, accrual, uploaded_at FROM orders WHERE number = \\$1").
		WithArgs("12345").
		WillReturnRows(sqlmock.NewRows([]string{"number", "username", "status", "accrual", "uploaded_at"}).
			AddRow("12345", "testuser", "PROCESSED", 100.0, time.Now()))

	ctx := context.Background()
	newOrder := models.Order{Number: "12345"}

	order, exists, err := repo.CreateOrder(ctx, newOrder)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "12345", order.Number)
	assert.Equal(t, "testuser", order.Username)
	assert.Equal(t, models.StatusProcessed, order.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersByUsername_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery("SELECT number, username, status, accrual, uploaded_at FROM orders WHERE username = \\$1").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"number", "username", "status", "accrual", "uploaded_at"}).
			AddRow("12345", "testuser", "NEW", 0.0, time.Now()).
			AddRow("67890", "testuser", "PROCESSED", 50.0, time.Now()))

	ctx := context.Background()
	orders, err := repo.GetOrdersByUsername(ctx, "testuser")

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, "12345", orders[0].Number)
	assert.Equal(t, "67890", orders[1].Number)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateOrders_Success(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	mock.ExpectBegin()

	mock.ExpectPrepare("UPDATE orders SET status = \\$1, accrual = \\$2 WHERE number = \\$3").
		ExpectExec().
		WithArgs("PROCESSED", 100.0, "12345").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	ctx := context.Background()
	orders := []models.Order{
		{Number: "12345", Status: "PROCESSED", Accrual: 100.0},
	}

	err := repo.UpdateOrders(ctx, orders)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWithdrawalsByUsername(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	username := "testuser"
	expectedOrders := []models.Order{
		{Number: "order1", Username: username, Status: "processed", Accrual: -50, UploadedAt: time.Now()},
		{Number: "order2", Username: username, Status: "processed", Accrual: -30, UploadedAt: time.Now().Add(-time.Hour)},
	}

	rows := sqlmock.NewRows([]string{"number", "username", "status", "accrual", "uploaded_at"}).
		AddRow(expectedOrders[0].Number, expectedOrders[0].Username, expectedOrders[0].Status, expectedOrders[0].Accrual, expectedOrders[0].UploadedAt).
		AddRow(expectedOrders[1].Number, expectedOrders[1].Username, expectedOrders[1].Status, expectedOrders[1].Accrual, expectedOrders[1].UploadedAt)

	mock.ExpectQuery("SELECT number, username, status, accrual, uploaded_at FROM orders WHERE username = \\$1 AND accrual < 0").
		WithArgs(username).
		WillReturnRows(rows)

	orders, err := repo.GetWithdrawalsByUsername(context.Background(), username)
	assert.NoError(t, err)
	assert.Len(t, orders, len(expectedOrders))

	for i, order := range orders {
		assert.Equal(t, expectedOrders[i].Number, order.Number)
		assert.Equal(t, expectedOrders[i].Username, order.Username)
		assert.Equal(t, expectedOrders[i].Status, order.Status)
		assert.Equal(t, math.Abs(expectedOrders[i].Accrual), order.Accrual) // Ensure conversion to positive
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersByStatus(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	expectedOrders := []models.Order{
		{Number: "order1", Username: "user1", Status: "pending", Accrual: 100, UploadedAt: time.Now()},
		{Number: "order2", Username: "user2", Status: "new", Accrual: 0, UploadedAt: time.Now().Add(-time.Hour)},
	}

	rows := sqlmock.NewRows([]string{"number", "username", "status", "accrual", "uploaded_at"}).
		AddRow(expectedOrders[0].Number, expectedOrders[0].Username, expectedOrders[0].Status, expectedOrders[0].Accrual, expectedOrders[0].UploadedAt).
		AddRow(expectedOrders[1].Number, expectedOrders[1].Username, expectedOrders[1].Status, expectedOrders[1].Accrual, expectedOrders[1].UploadedAt)

	mock.ExpectQuery("SELECT number, username, status, accrual, uploaded_at FROM orders WHERE status <> \\$1").
		WithArgs(models.StatusProcessed).
		WillReturnRows(rows)

	orders, err := repo.GetOrdersByStatus(context.Background())
	assert.NoError(t, err)
	assert.Len(t, orders, len(expectedOrders))

	for i, order := range orders {
		assert.Equal(t, expectedOrders[i].Number, order.Number)
		assert.Equal(t, expectedOrders[i].Username, order.Username)
		assert.Equal(t, expectedOrders[i].Status, order.Status)
		assert.Equal(t, expectedOrders[i].Accrual, order.Accrual)
		assert.Equal(t, expectedOrders[i].UploadedAt, order.UploadedAt)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserBalanceAndWithdrawals(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	username := "testuser"
	expectedBalance := 150.0
	expectedWithdrawals := -50.0 // Stored as negative in DB

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(accrual\\), 0\\), COALESCE\\(SUM\\(CASE WHEN accrual < 0 THEN accrual ELSE 0 END\\), 0\\) FROM orders WHERE username = \\$1").
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "withdrawals"}).AddRow(expectedBalance, expectedWithdrawals))

	balance, withdrawals, err := repo.GetUserBalanceAndWithdrawals(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	assert.Equal(t, -expectedWithdrawals, withdrawals) // Convert to positive

	assert.NoError(t, mock.ExpectationsWereMet())
}

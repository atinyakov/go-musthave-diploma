package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/internal/db"
)

type Repository struct {
	db *sql.DB
}

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(username, password string) error {
	// Check if the user already exists
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return ErrUserExists
	}

	// Insert new user
	_, err = r.db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, password)
	return err
}

func (r *Repository) Login(username string) (string, error) {
	var hashedPassword string
	err := db.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		return "", ErrUserNotFound // User doesn't exist
	}
	if err != nil {
		return "", err // Other errors
	}
	return hashedPassword, nil
}

func (r *Repository) CreateOrder(ctx context.Context, newOrder models.Order) (*models.Order, bool, error) {
	// Check if the order already exists
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)", newOrder.Number).Scan(&exists)
	if err != nil {
		return nil, false, err
	}

	if exists {
		// If the order exists, fetch the existing order details
		var order models.Order
		err := r.db.QueryRowContext(ctx, "SELECT number, username, status, accrual, uploaded_at FROM orders WHERE number = $1", newOrder.Number).
			Scan(&order.Number, &order.Username, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, true, err
		}
		return &order, true, nil
	}

	// If the order doesn't exist, insert a new order into the `orders` table
	_, err = r.db.ExecContext(ctx, "INSERT INTO orders (number, username, status, accrual) VALUES ($1, $2, $3, $4)", newOrder.Number, newOrder.Username, newOrder.Status, newOrder.Accrual)
	if err != nil {
		return nil, false, err
	}

	// Return the newly created order
	createdOrder := &models.Order{
		Username:   newOrder.Username,
		Number:     newOrder.Number,
		Status:     newOrder.Status,
		UploadedAt: time.Now(), // Set the uploaded_at field to the current time
	}

	return createdOrder, false, nil
}

func (r *Repository) GetOrdersByStatus(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT number, username, status, accrual, uploaded_at FROM orders WHERE status <> $1", models.StatusProcessed)
	if err != nil {
		fmt.Printf("GetOrdersByStatus error=%s", err.Error())
		return []models.Order{}, nil
	}
	defer rows.Close()

	res := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.Username, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			fmt.Printf("GetOrdersByStatus error=%s", err.Error())
			return nil, err
		}

		res = append(res, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil

}

func (r *Repository) GetOrdersByUsername(ctx context.Context, username string) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT number, username, status, accrual, uploaded_at FROM orders WHERE username = $1", username)
	if err != nil {
		fmt.Printf("GetOrdersByUsername error=%s", err.Error())
		return []models.Order{}, nil
	}
	defer rows.Close()

	res := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.Username, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			fmt.Printf("GetOrdersByUsername error=%s", err.Error())
			return nil, err
		}

		res = append(res, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) UpdateOrders(ctx context.Context, os []models.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(`
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3;
	`)
	if err != nil {
		return err
	}

	defer stmt.Close()
	for _, o := range os {

		_, err = stmt.ExecContext(ctx, o.Status, o.Accrual, o.Number)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetWithdrawalsByUsername(ctx context.Context, username string) ([]models.Order, error) {
	query := `
	SELECT number, username, status, accrual, uploaded_at 
	FROM orders 
	WHERE username = $1 AND accrual < 0 
	ORDER BY uploaded_at DESC;`

	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		fmt.Printf("GetWithdrawalsByUsername error=%s", err.Error())
		return []models.Order{}, nil
	}
	defer rows.Close()

	res := make([]models.Order, 0)

	for rows.Next() {
		var order models.Order

		err := rows.Scan(&order.Number, &order.Username, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			fmt.Printf("GetWithdrawalsByUsername error=%s", err.Error())
			return nil, err
		}
		// Convert negative accrual to positive
		order.Accrual = math.Abs(order.Accrual) // Assuming Accrual is a float
		// If Accrual is an int, use: order.Accrual = -order.Accrual

		res = append(res, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *Repository) GetUserBalanceAndWithdrawals(ctx context.Context, username string) (float64, float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(accrual), 0), 
			COALESCE(SUM(CASE WHEN accrual < 0 THEN accrual ELSE 0 END), 0) 
		FROM orders 
		WHERE username = $1;`

	var balance, withdrawals float64
	err := r.db.QueryRowContext(ctx, query, username).Scan(&balance, &withdrawals)
	if err != nil {
		fmt.Printf("GetUserBalanceAndWithdrawals error: %s\n", err.Error())
		return 0, 0, err
	}

	// Convert withdrawals to positive since they are stored as negative values
	return balance, -withdrawals, nil
}

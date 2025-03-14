package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r *Repository) CreateOrder(orderNumber int, username string, status models.OrderStatus) (*models.Order, bool, error) {
	// Check if the order already exists
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1)", orderNumber).Scan(&exists)
	if err != nil {
		return nil, false, err
	}

	if exists {
		// If the order exists, fetch the existing order details
		var order models.Order
		err := r.db.QueryRow("SELECT number, username, status, accrual, uploaded_at FROM orders WHERE number = $1", orderNumber).
			Scan(&order.Number, &order.Username, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, true, err
		}
		return &order, true, nil
	}

	// If the order doesn't exist, insert a new order into the `orders` table
	_, err = r.db.Exec("INSERT INTO orders (number, username, status) VALUES ($1, $2, $3)", orderNumber, username, status)
	if err != nil {
		return nil, false, err
	}

	// Return the newly created order
	newOrder := &models.Order{
		Username:   username,
		Number:     fmt.Sprintf("%d", orderNumber),
		Status:     status,
		UploadedAt: time.Now(), // Set the uploaded_at field to the current time
	}

	return newOrder, false, nil
}

func (r *Repository) GetOrdersByStatus(ctx context.Context, status models.OrderStatus) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT number, username, status, accrual, uploaded_at FROM orders WHERE status = $1", status)
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
			return nil, nil
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
			fmt.Printf("GetOrdersByStatus error=%s", err.Error())
			return nil, nil
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
		err := tx.Rollback()
		if err != nil {
			// r.logger.Error("ROLLBACK error=", zap.String("error", err.Error()))
			// return err
			fmt.Println(err.Error())
		}
	}()

	stmt, err := tx.Prepare(`
		UPDATE orders
		SET status = $1
		WHERE username = $2;
	`)
	if err != nil {
		return err
	}

	for _, o := range os {

		defer stmt.Close()
		_, err = stmt.ExecContext(ctx, o.Status, o.Username)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetOrdersByID() {
	// return true, nil
}

func (r *Repository) GetWithdrawalsById() {
	// return true, nil
}

func (r *Repository) GetBalance(string, string) (bool, error) {
	return true, nil
}

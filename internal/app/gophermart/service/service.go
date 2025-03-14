package service

import (
	"context"
	"errors"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/pkg/auth"
	"github.com/atinyakov/go-musthave-diploma/pkg/luhn"
)

type Repository interface {
	CreateUser(string, string) error
	Login(string) (string, error)
	CreateOrder(orderNumber int, username string, status models.OrderStatus) (*models.Order, bool, error)
	GetOrdersByUsername(ctx context.Context, username string) ([]models.Order, error)
	GetWithdrawalsByID()
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

var ErrInvalidLuhn = errors.New("order is invalid")
var ErrExists = errors.New("order already exists")
var ErrNotBelongsToUser = errors.New("order is created by another user")

func (r *Service) Register(login string, password string) error {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	err = r.repo.CreateUser(login, hashedPassword)

	if err == repository.ErrUserExists {
		return err // Pass it to handler
	} else if err != nil {
		return errors.New("failed to create user")
	}

	return nil
}

func (r *Service) Login(login string, password string) (bool, error) {
	hashedPassword, err := r.repo.Login(login)
	if err == repository.ErrUserNotFound {
		return false, errors.New("invalid username or password")
	}
	if err != nil {
		return false, err
	}

	isValid := auth.CheckPassword(hashedPassword, password)

	if !isValid {
		return false, errors.New("invalid username or password")
	}
	return true, nil
}

func (r *Service) CreateOrder(orderNumber int, username string) error {
	isValid := luhn.Valid(orderNumber)

	if !isValid {
		return ErrInvalidLuhn
	}

	order, exists, err := r.repo.CreateOrder(orderNumber, username, models.StatusNew)

	if exists {
		if order.Username == username {
			return ErrExists
		}
		if order.Username != username {
			return ErrNotBelongsToUser
		}
	}

	if err != nil {
		return err
	}

	// TODO: sent to accuralservice

	return nil
}

func (r *Service) GetOrdersByUsername(ctx context.Context, username string) ([]models.Order, error) {
	return r.repo.GetOrdersByUsername(ctx, username)
}

func (r *Service) GetWithdrawals(string, string) (bool, error) {
	return true, nil
}

func (r *Service) GetBalance(string, string) (bool, error) {
	return true, nil
}

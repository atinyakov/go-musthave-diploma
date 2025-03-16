package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/pkg/auth"
	"github.com/atinyakov/go-musthave-diploma/pkg/luhn"
)

type Repository interface {
	CreateUser(string, string) error
	Login(string) (string, error)
	CreateOrder(context.Context, models.Order) (*models.Order, bool, error)
	GetOrdersByUsername(ctx context.Context, username string) ([]models.Order, error)
	GetWithdrawalsByUsername(ctx context.Context, username string) ([]models.Order, error)
	GetUserBalanceAndWithdrawals(ctx context.Context, username string) (float64, float64, error)
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
		return err
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	order, exists, err := r.repo.CreateOrder(ctx, models.Order{Number: strconv.Itoa(orderNumber), Username: username, Status: models.StatusNew, Accrual: 0})

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

	return nil
}

func (r *Service) GetOrdersByUsername(username string) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.repo.GetOrdersByUsername(ctx, username)
}

func (r *Service) CreateWidthraw(req dto.WithdrawalRequest, username string) error {
	order, err := strconv.Atoi(req.Order)
	if err != nil {
		return err
	}

	isValid := luhn.Valid(order)

	if !isValid {
		return ErrInvalidLuhn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, _, err = r.repo.CreateOrder(ctx, models.Order{Number: req.Order, Username: username, Accrual: -float64(req.Sum)})

	return err
}

func (r *Service) GetBalance(username string) (dto.BalanceResponce, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	balance, widthdraw, err := r.repo.GetUserBalanceAndWithdrawals(ctx, username)
	if err != nil {
		return dto.BalanceResponce{}, err
	}

	return dto.BalanceResponce{Current: float32(RoundTo(balance, 2)), Withdrawn: float32(RoundTo(widthdraw, 2))}, nil
}

func (r *Service) GetWithdrawals(username string) ([]dto.WithdrawalResponseItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	widthdrawals, err := r.repo.GetWithdrawalsByUsername(ctx, username)

	if err != nil {
		return []dto.WithdrawalResponseItem{}, err
	}

	var res []dto.WithdrawalResponseItem

	for _, w := range widthdrawals {
		res = append(res, dto.WithdrawalResponseItem{ProcessedAt: w.UploadedAt, Order: w.Number, Sum: RoundTo(w.Accrual, 2)})
	}

	return res, nil
}

func RoundTo(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

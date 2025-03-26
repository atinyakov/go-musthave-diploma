package service_test

import (
	"strconv"
	"testing"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/mocks"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/service"
	"github.com/atinyakov/go-musthave-diploma/pkg/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	srv := service.New(mockRepo)

	login := "testuser"
	password := "testpassword"
	// hashedPassword, _ := auth.HashPassword(password)

	// Use gomock.Any() to match any password value
	mockRepo.EXPECT().CreateUser(login, gomock.Any()).Return(nil)

	// Test successful registration
	err := srv.Register(login, password)
	assert.NoError(t, err)

	// Test user already exists error
	mockRepo.EXPECT().CreateUser(login, gomock.Any()).Return(repository.ErrUserExists)
	err = srv.Register(login, password)
	assert.EqualError(t, err, repository.ErrUserExists.Error())
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	srv := service.New(mockRepo)

	login := "testuser"
	password := "testpassword"
	hashedPassword, _ := auth.HashPassword(password)

	// Case where user exists and password matches
	mockRepo.EXPECT().GetPasswordHashByUsername(login).Return(hashedPassword, nil)
	valid, err := srv.Login(login, password)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Case where user doesn't exist
	mockRepo.EXPECT().GetPasswordHashByUsername(login).Return("", repository.ErrUserNotFound)
	valid, err = srv.Login(login, password)
	assert.EqualError(t, err, "invalid username or password")
	assert.False(t, valid)
}

func TestCreateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	srv := service.New(mockRepo)

	orderNumber := 4012888888881881
	username := "testuser"
	order := models.Order{
		Number:   strconv.Itoa(orderNumber),
		Username: username,
		Status:   models.StatusNew,
		Accrual:  0,
	}

	mockRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Eq(order)).Return(&order, false, nil)

	err := srv.CreateOrder(orderNumber, username)
	assert.NoError(t, err)

	mockRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Eq(order)).Return(&order, true, nil)
	err = srv.CreateOrder(orderNumber, username)
	assert.Equal(t, service.ErrExists, err)
}

func TestGetOrdersByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.New(mockRepo)

	username := "testuser"
	expectedOrders := []models.Order{
		{Number: "123", Username: username, Status: models.StatusNew},
	}

	mockRepo.EXPECT().GetOrdersByUsername(gomock.Any(), username).Return(expectedOrders, nil)

	orders, err := service.GetOrdersByUsername(username)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
}

func TestCreateWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	srv := service.New(mockRepo)

	username := "testuser"
	req := dto.WithdrawalRequest{Order: "123", Sum: 100}
	validReq := dto.WithdrawalRequest{Order: "4012888888881881", Sum: 100}

	// Test invalid Luhn number
	err := srv.CreateWidthraw(req, username)
	assert.EqualError(t, err, service.ErrInvalidLuhn.Error())

	// Test valid withdrawal
	mockRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(&models.Order{}, false, nil)
	err = srv.CreateWidthraw(validReq, username)
	assert.NoError(t, err)
}

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.New(mockRepo)

	username := "testuser"
	expectedBalance := 100.0
	expectedWithdrawals := 50.0

	mockRepo.EXPECT().GetUserBalanceAndWithdrawals(gomock.Any(), username).Return(expectedBalance, expectedWithdrawals, nil)

	balance, err := service.GetBalance(username)
	assert.NoError(t, err)
	assert.Equal(t, float32(100.0), balance.Current)
	assert.Equal(t, float32(50.0), balance.Withdrawn)
}

func TestGetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.New(mockRepo)

	username := "testuser"
	expectedWithdrawals := []models.Order{
		{Number: "123", Username: username, Accrual: -100.0},
	}

	mockRepo.EXPECT().GetWithdrawalsByUsername(gomock.Any(), username).Return(expectedWithdrawals, nil)

	withdrawals, err := service.GetWithdrawals(username)
	assert.NoError(t, err)
	assert.Len(t, withdrawals, 1)
	assert.Equal(t, "123", withdrawals[0].Order)
	assert.Equal(t, -100.0, withdrawals[0].Sum)
}

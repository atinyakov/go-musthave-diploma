package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/handler"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/mocks"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/service"
	"github.com/atinyakov/go-musthave-diploma/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServicePost(ctrl)
	h := handler.NewPost(mockService)

	reqData := dto.UserRequest{Login: "testuser", Password: "password"}
	reqBody, _ := json.Marshal(reqData)

	t.Run("success", func(t *testing.T) {
		mockService.EXPECT().Register(reqData.Login, reqData.Password).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Authorization"), "Bearer")
	})

	t.Run("user exists", func(t *testing.T) {
		mockService.EXPECT().Register(reqData.Login, reqData.Password).Return(repository.ErrUserExists)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
	t.Run("err", func(t *testing.T) {
		mockService.EXPECT().Register(reqData.Login, reqData.Password).Return(errors.New("123"))

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServicePost(ctrl)
	h := handler.NewPost(mockService)

	t.Run("valid login", func(t *testing.T) {
		mockService.EXPECT().Login("testuser", "password").Return(true, nil)
		reqData := dto.UserRequest{Login: "testuser", Password: "password"}
		reqBody, _ := json.Marshal(reqData)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Authorization"), "Bearer")
	})
	t.Run("invalid login", func(t *testing.T) {
		mockService.EXPECT().Login("testuser", "password").Return(false, errors.New("invalid username or password"))
		reqData := dto.UserRequest{Login: "testuser", Password: "password"}
		reqBody, _ := json.Marshal(reqData)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("err", func(t *testing.T) {
		mockService.EXPECT().Login("testuser", "password").Return(false, errors.New("123"))
		reqData := dto.UserRequest{Login: "testuser", Password: "password"}
		reqBody, _ := json.Marshal(reqData)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.Login(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPostOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServicePost(ctrl)
	h := handler.NewPost(mockService)

	t.Run("valid order", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(12345, "testuser").Return(nil)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("12345")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})
	t.Run("invalid Luhn", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(12345, "testuser").Return(service.ErrInvalidLuhn)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("12345")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
	t.Run("invalid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("123b45")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("order exists", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(12345, "testuser").Return(service.ErrExists)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("12345")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("order not belong to user", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(12345, "testuser").Return(service.ErrNotBelongsToUser)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("12345")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
	t.Run("error", func(t *testing.T) {
		mockService.EXPECT().CreateOrder(12345, "testuser").Return(errors.New("123"))
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer([]byte("12345")))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestBalanceWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServicePost(ctrl)
	h := handler.NewPost(mockService)

	t.Run("valid withdraw", func(t *testing.T) {
		withdrawReq := dto.WithdrawalRequest{Order: "12345", Sum: 100}
		mockService.EXPECT().CreateWidthraw(withdrawReq, "testuser").Return(nil)
		reqBody, _ := json.Marshal(withdrawReq)

		req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.BalanceWithdraw(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid withdraw", func(t *testing.T) {
		withdrawReq := dto.WithdrawalRequest{Order: "12345", Sum: 100}
		mockService.EXPECT().CreateWidthraw(withdrawReq, "testuser").Return(errors.New("123"))
		reqBody, _ := json.Marshal(withdrawReq)

		req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.BalanceWithdraw(w, req)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
	t.Run("invalid req", func(t *testing.T) {
		withdrawReq := dto.BalanceResponce{Current: 12345, Withdrawn: 100}
		reqBody, _ := json.Marshal(withdrawReq)

		req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.BalanceWithdraw(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

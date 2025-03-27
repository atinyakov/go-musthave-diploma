package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/handler"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/mocks"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/pkg/middleware"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestGetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceGet(ctrl)
	h := handler.NewGet(mockService)

	t.Run("success", func(t *testing.T) {
		mockService.EXPECT().GetOrdersByUsername("testuser").Return([]models.Order{{Number: "12345"}}, nil)
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Orders(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceGet(ctrl)
	h := handler.NewGet(mockService)

	t.Run("success", func(t *testing.T) {
		mockService.EXPECT().GetBalance("testuser").Return(dto.BalanceResponce{Current: 100.0}, nil)
		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Balance(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockServiceGet(ctrl)
	h := handler.NewGet(mockService)

	t.Run("success", func(t *testing.T) {
		mockService.EXPECT().GetWithdrawals("testuser").Return([]dto.WithdrawalResponseItem{{Order: "12345", Sum: 50.0}}, nil)
		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, "testuser"))
		w := httptest.NewRecorder()

		h.Withdrawals(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"github.com/atinyakov/go-musthave-diploma/pkg/middleware"
)

type ServiceGet interface {
	GetWithdrawals(string) ([]dto.WithdrawalResponseItem, error)
	GetBalance(string) (dto.BalanceResponce, error)
	GetOrdersByUsername(string) ([]models.Order, error)
}

type GetHandler struct {
	service ServiceGet
}

func NewGet(service ServiceGet) *GetHandler {
	return &GetHandler{
		service: service,
	}
}

func (gh *GetHandler) Orders(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserContextKey).(string)

	orders, err := gh.service.GetOrdersByUsername(username)

	if err != nil {
		slog.Error("Get Orders DB error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		slog.Error("Get Orders Marshal error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		slog.Error("writeErr error", slog.String("error", writeErr.Error()))
	}
}

func (gh *GetHandler) Balance(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserContextKey).(string)

	balance, err := gh.service.GetBalance(username)

	if err != nil {
		slog.Error("Get Balance DB error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(balance)
	if err != nil {
		slog.Error("Get Balance Marshal error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		slog.Error("writeErr error", slog.String("error", writeErr.Error()))
	}
}

func (gh *GetHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserContextKey).(string)

	widthdrawals, err := gh.service.GetWithdrawals(username)

	if err != nil {
		slog.Error("Get Withdrawals DB error", slog.String("error", err.Error()))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(widthdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(widthdrawals)
	if err != nil {
		slog.Error("Get Withdrawals Marshal error", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		slog.Error("writeErr error", slog.String("error", writeErr.Error()))
	}
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
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
	username := r.Header.Get("X-User")

	orders, err := gh.service.GetOrdersByUsername(username)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (gh *GetHandler) Balance(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")

	balance, err := gh.service.GetBalance(username)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	response, err := json.Marshal(balance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (gh *GetHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-User")

	widthdrawals, err := gh.service.GetWithdrawals(username)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(widthdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(widthdrawals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(response)
	if writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

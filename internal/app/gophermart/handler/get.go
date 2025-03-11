package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
)

type ServiceGet interface {
	GetWithdrawals(string, string) (bool, error)
	GetBalance(string, string) (bool, error)
	GetOrdersByUsername(context.Context, string) ([]models.Order, error)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	orders, err := gh.service.GetOrdersByUsername(ctx, username)

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

}

func (gh *GetHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {

}

package handler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/service"
	"github.com/atinyakov/go-musthave-diploma/pkg/auth"
	"github.com/atinyakov/go-musthave-diploma/pkg/middleware"
)

type ServicePost interface {
	Register(string, string) error
	Login(string, string) (bool, error)
	CreateOrder(int, string) error
	CreateWidthraw(dto.WithdrawalRequest, string) error
}

type PostHandler struct {
	service ServicePost
}

func NewPost(service ServicePost) *PostHandler {
	return &PostHandler{
		service: service,
	}
}

func (ph *PostHandler) Register(w http.ResponseWriter, r *http.Request) {
	var reqData dto.UserRequest

	err := decodeJSONBody(w, r, &reqData)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
			return
		}

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	err = ph.service.Register(reqData.Login, reqData.Password)

	if errors.Is(err, repository.ErrUserExists) {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWT(reqData.Login)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)

}

func (ph *PostHandler) Login(w http.ResponseWriter, r *http.Request) {
	var reqData dto.UserRequest

	err := decodeJSONBody(w, r, &reqData)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	isValid, err := ph.service.Login(reqData.Login, reqData.Password)
	if err != nil {
		if err.Error() == "invalid username or password" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if isValid {
		// Generate JWT
		token, err := auth.GenerateJWT(reqData.Login)
		if err != nil {
			slog.Error("Error generating token", slog.String("error", err.Error()))

			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Respond with token
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Authorization", "Bearer "+token)

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func (ph *PostHandler) Orders(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserContextKey).(string)

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	err = ph.service.CreateOrder(orderNumber, username)

	if errors.Is(err, service.ErrInvalidLuhn) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return

	}

	if errors.Is(err, service.ErrExists) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, service.ErrNotBelongsToUser) {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)

		return
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (ph *PostHandler) BalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UserContextKey).(string)

	var reqData dto.WithdrawalRequest

	err := decodeJSONBody(w, r, &reqData)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			slog.Error(mr.msg, slog.String("status", strconv.Itoa(mr.status)))
			http.Error(w, mr.msg, mr.status)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	err = ph.service.CreateWidthraw(reqData, username)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("error balance widtdraw: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

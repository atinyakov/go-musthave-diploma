package server

import (
	"net/http"
	"time"

	authMiddleware "github.com/atinyakov/go-musthave-diploma/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type PostHandler interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
	Orders(http.ResponseWriter, *http.Request)
	BalanceWithdraw(http.ResponseWriter, *http.Request)
}

type GetHandler interface {
	Balance(http.ResponseWriter, *http.Request)
	Orders(http.ResponseWriter, *http.Request)
	Withdrawals(http.ResponseWriter, *http.Request)
}

func New(post PostHandler, get GetHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json", "text/plain"))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hi"))
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", post.Register)
		r.Post("/login", post.Login)

		// Secured Routes
		r.With(authMiddleware.AuthMiddleware).Post("/orders", post.Orders)
		r.With(authMiddleware.AuthMiddleware).Get("/orders", get.Orders)

		r.With(authMiddleware.AuthMiddleware).Get("/balance", get.Balance)
		r.With(authMiddleware.AuthMiddleware).Post("/balance/withdraw", post.BalanceWithdraw)

		r.With(authMiddleware.AuthMiddleware).Get("/withdrawals", get.Withdrawals)
	})

	return r
}

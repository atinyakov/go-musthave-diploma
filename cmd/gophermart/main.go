package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/atinyakov/go-musthave-diploma/internal/app/config"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/client"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/handler"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/server"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/service"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/worker"
	"github.com/atinyakov/go-musthave-diploma/internal/db"
)

func main() {
	config := config.LoadConfig()
	db := db.InitDB(config.DatabaseURI)
	client := client.New(config.AccrualSystemAddress + "/api/orders/")

	repository := repository.New(db)
	worker := worker.NewAccrualTaskWorker(repository, client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.StartOrderFetcher(ctx)

	slog.Info("workers created")

	service := service.New(repository)
	postHandler := handler.NewPost(service)
	getHandler := handler.NewGet(service)
	r := server.New(postHandler, getHandler)
	slog.Info("Starting server")

	err := http.ListenAndServe(config.RunAddress, r)
	if err != nil {
		panic(err)
	}
}

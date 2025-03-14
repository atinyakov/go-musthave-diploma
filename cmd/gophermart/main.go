package main

import (
	"fmt"
	"net/http"

	// "os/exec"

	// "runtime"

	"github.com/atinyakov/go-musthave-diploma/internal/app/config"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/handler"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/repository"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/server"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/service"
	"github.com/atinyakov/go-musthave-diploma/internal/db"
)

func main() {
	config := config.LoadConfig()
	db := db.InitDB(config.DatabaseURI)
	fmt.Println(config.AccrualSystemAddress + "/api/orders/")
	// client := client.New(config.AccrualSystemAddress + "/api/orders/")

	repository := repository.New(db)
	// worker := worker.NewAccrualTaskWorker(repository, client)
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel() // Ensures cleanup on exit
	// defer cancel()
	// go worker.StartOrderFetcher(ctx)
	// go worker.StartOrderUpdater(ctx, 10)
	// fmt.Println("workers created")

	service := service.New(repository)
	postHandler := handler.NewPost(service)
	getHandler := handler.NewGet(service)
	r := server.New(postHandler, getHandler)
	fmt.Println("Starting server")

	err := http.ListenAndServe(config.RunAddress, r)
	fmt.Println("Running server")
	if err != nil {
		panic(err)
	}
}

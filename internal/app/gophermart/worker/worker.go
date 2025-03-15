package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
)

type Repo interface {
	UpdateOrders(ctx context.Context, os []models.Order) error
	GetOrdersByStatus(ctx context.Context, status models.OrderStatus) ([]models.Order, error)
}

type Client interface {
	Request(url string) (*models.Order, error)
}

type AccrualTaskWorker struct {
	in     chan models.Order
	client Client
	repo   Repo
}

func NewAccrualTaskWorker(repo Repo, client Client) *AccrualTaskWorker {
	ch := make(chan models.Order, 100) // Buffered channel to avoid blocking
	return &AccrualTaskWorker{
		in:     ch,
		client: client,
		repo:   repo,
	}
}

func (s *AccrualTaskWorker) StartOrderFetcher(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("StartOrderFetcher shutting down...")
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			fmt.Println("StartOrderFetcher getting orders...")

			newOrders, err := s.repo.GetOrdersByStatus(ctx, models.StatusNew)
			if err != nil {
				fmt.Printf("Failed to fetch new orders: %s\n", err)
				continue
			}

			if len(newOrders) == 0 {
				continue
			}

			for _, order := range newOrders {
				s.in <- order
			}
		}
	}
}

func (s *AccrualTaskWorker) StartOrderUpdater(ctx context.Context, numWorkers int) {
	var wg sync.WaitGroup

	// Worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// var messages []models.Order

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Worker shutting down...")
					return
				case msg := <-s.in:
					fmt.Println("Worker GOT NEW ORDER...")

					_, err := s.client.Request(msg.Number)
					if err != nil {
						fmt.Printf("Failed to update order %s: %s\n", msg.Number, err)
						continue
					}

					// if orderResponse != nil {
					// 	messages = append(messages, *orderResponse)
					// }
				}
			}
		}()
	}

	// Main routine to update orders
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var messages []models.Order
	for {
		select {
		case <-ctx.Done():
			fmt.Println("OrderUpdater shutting down...")
			return
		case <-ticker.C:
			if len(messages) > 0 {
				fmt.Println("UPDATING ORDERS BACK TO DB")
				err := s.repo.UpdateOrders(ctx, messages)
				if err != nil {
					fmt.Printf("Failed to update orders: %s\n", err)
				}
				messages = messages[:0] // Reset the messages slice
			} else {
				continue
			}
		}
	}
}

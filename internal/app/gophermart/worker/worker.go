package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
)

type Repo interface {
	UpdateOrders(ctx context.Context, os []models.Order) error
	GetOrdersByStatus(ctx context.Context) ([]models.Order, error)
}

type Client interface {
	Request(url string) (*models.Order, error)
}

type AccrualTaskWorker struct {
	client Client
	repo   Repo
}

func NewAccrualTaskWorker(repo Repo, client Client) *AccrualTaskWorker {
	return &AccrualTaskWorker{
		client: client,
		repo:   repo,
	}
}

func (s *AccrualTaskWorker) StartOrderFetcher(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
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

			newOrders, err := s.repo.GetOrdersByStatus(ctx)
			if err != nil {
				fmt.Printf("Failed to fetch new orders: %s\n", err)
				continue
			}

			if len(newOrders) == 0 {
				continue
			}

			for _, order := range newOrders {
				res, err := s.client.Request(order.Number)
				if err != nil {
					fmt.Printf("Failed to update order %s: %s\n", order.Number, err)
					return
				}

				if res != nil {
					err = s.repo.UpdateOrders(ctx, append([]models.Order{}, *res))
					if err != nil {
						fmt.Printf("Failed to update orders: %s\n", err)
					}
				}
			}

		}
	}
}

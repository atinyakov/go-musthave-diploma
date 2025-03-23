package worker

import (
	"context"
	"errors"
	"time"

	"github.com/gookit/slog"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/client"
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
			slog.Info("StartOrderFetcher shutting down...")
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			slog.Info("StartOrderFetcher getting orders...")

			newOrders, err := s.repo.GetOrdersByStatus(ctx)
			if err != nil {
				slog.Error("Failed to fetch new orders: %s\n", err)
				continue
			}

			if len(newOrders) == 0 {
				continue
			}

			for _, order := range newOrders {
				res, err := s.client.Request(order.Number)

				var rae client.RetryAfterErr
				if errors.As(err, &rae) {
					slog.Info("sleeping %d seconds\n", rae.T)
					time.Sleep(time.Duration(rae.T * int(time.Second)))
				}

				if res != nil {
					err = s.repo.UpdateOrders(ctx, append([]models.Order{}, *res))
					if err != nil {
						slog.Error("Failed to update orders: %s\n", err)
					}
				}
			}

		}
	}
}

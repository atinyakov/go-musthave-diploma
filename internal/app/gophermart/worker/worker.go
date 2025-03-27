package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/client"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
)

type Repo interface {
	UpdateOrders(ctx context.Context, os []models.Order) error
	GetOrdersByStatus(ctx context.Context) ([]models.Order, error)
}

type Client interface {
	Request(context.Context, string) (*models.Order, error)
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
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			slog.Info("StartOrderFetcher getting orders...")

			newOrders, err := s.repo.GetOrdersByStatus(ctx)
			if err != nil {
				slog.Error("Failed to fetch new orders: %s\n", slog.String("error", err.Error()))
				continue
			}

			if len(newOrders) == 0 {
				continue
			}

			for _, order := range newOrders {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				res, err := s.client.Request(ctx, order.Number)

				var rae client.RetryAfterErr
				if errors.As(err, &rae) {
					slog.Info("sleeping %d seconds\n", slog.Int("time", rae.T))
					time.Sleep(time.Duration(rae.T * int(time.Second)))
				}

				if res != nil {
					err = s.repo.UpdateOrders(ctx, append([]models.Order{}, *res))
					if err != nil {
						slog.Error("Failed to update orders: %s\n", slog.String("error", err.Error()))
					}
				}
			}

		}
	}
}

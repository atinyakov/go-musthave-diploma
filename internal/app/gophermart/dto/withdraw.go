package dto

import (
	"time"
)

// WithdrawalRequest represents a withdrawal request.
type WithdrawalRequest struct {
	Order string  `json:"order" validate:"required"`
	Sum   float64 `json:"sum" validate:"required,gt=0"`
}

// WithdrawalResponseItem represents a processed withdrawal response.
type WithdrawalResponseItem struct {
	Order       string    `json:"order" validate:"required"`
	Sum         float64   `json:"sum" validate:"required,gt=0"`
	ProcessedAt time.Time `json:"processed_at" format:"RFC3339"`
}

package dto

import "time"

type WithdrawalRequest struct {
	Order int `json:"order" validate:"required"`
	Sum   int `json:"sum" validate:"required"`
}

type WithdrawalResponceItem struct {
	Order       int       `json:"order" validate:"required"`
	Sum         int       `json:"sum" validate:"required"`
	ProcessedAt time.Time `json:"processed_at" ` //RFC3339
}

package dto

type BalanceResponce struct {
	Current   float32 `json:"current" validate:"required"`
	Withdrawn float32 `json:"withdrawn" validate:"required"`
}

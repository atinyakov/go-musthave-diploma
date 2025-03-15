package dto

type BalanceResponce struct {
	Current   float32 `json:"current" validate:"required"`
	Withdrawn int     `json:"withdrawn" validate:"required"`
}

package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
)

type AccrualClient struct {
	cli     *http.Client
	baseURL string
}

func New(baseURL string) *AccrualClient {
	return &AccrualClient{
		cli:     &http.Client{},
		baseURL: baseURL,
	}
}

func (c *AccrualClient) Request(url string) (*models.Order, error) {

	request, err := http.NewRequest(http.MethodGet, c.baseURL+url, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := c.cli.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	var accrual dto.AccrualResponse
	dec := json.NewDecoder(response.Body)
	dec.DisallowUnknownFields()

	err = dec.Decode(&accrual)
	if err != nil {
		return nil, err
	}

	return &models.Order{Accrual: accrual.Accrual, Status: models.OrderStatus(accrual.Status), Number: accrual.Order}, err
}

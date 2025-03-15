package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/dto"
	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
	retry "gopkg.in/h2non/gentleman-retry.v2"
	"gopkg.in/h2non/gentleman.v2"
)

type AccrualClient struct {
	cli     *gentleman.Client
	baseURL string
}

func New(baseURL string) *AccrualClient {
	// Create a new client
	cli := gentleman.New()

	// Define base URL
	cli.URL(baseURL)

	// Register the retry plugin, using a custom exponential retry strategy
	cli.Use(retry.New(retrier.New(retrier.ExponentialBackoff(3, 100*time.Millisecond), nil)))

	return &AccrualClient{
		cli:     cli,
		baseURL: baseURL,
	}
}

type ErrorResponse struct {
	Message string `json:"message"` // Adjust field based on actual API response
	Code    int    `json:"code,omitempty"`
}

func (c *AccrualClient) Request(url string) (*models.Order, error) {
	// Create a new request based on the current client
	// c.cli.
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodGet, c.baseURL+url, nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/json")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// // и печатаем его
	// fmt.Println(string(body))

	var accrual dto.AccrualResponse
	dec := json.NewDecoder(response.Body)
	dec.DisallowUnknownFields()

	err = dec.Decode(&accrual)
	if err != nil {
		return nil, err
	}

	return &models.Order{Accrual: accrual.Accrual, Status: models.OrderStatus(accrual.Status), Number: accrual.Order}, err
}

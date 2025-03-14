package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/atinyakov/go-musthave-diploma/internal/app/gophermart/models"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
	retry "gopkg.in/h2non/gentleman-retry.v2"
	"gopkg.in/h2non/gentleman.v2"
)

type AccrualClient struct {
	cli *gentleman.Client
}

func New(baseURL string) *AccrualClient {
	// Create a new client
	cli := gentleman.New()

	// Define base URL
	cli.URL(baseURL)

	// Register the retry plugin, using a custom exponential retry strategy
	cli.Use(retry.New(retrier.New(retrier.ExponentialBackoff(3, 100*time.Millisecond), nil)))

	return &AccrualClient{
		cli: cli,
	}
}

func (c *AccrualClient) Request(url string) (*models.Order, error) {
	// Create a new request based on the current client
	// c.cli.
	req := c.cli.Request()

	// Define the URL path at request level
	req.Path(url)

	// Set a new header field
	req.SetHeader("Client", "gentleman")
	req.SetHeader("Method", "GET")

	fmt.Println("sedning new request")
	res, err := req.Send()
	if err != nil {
		fmt.Printf("Request error: %s\n", err)
		return nil, err
	}
	if !res.Ok {
		fmt.Printf("Invalid server response: %d\n", res.StatusCode)
		return nil, err
	}
	fmt.Println("Got responce")

	var order models.Order
	dec := json.NewDecoder(res.RawResponse.Body)
	dec.DisallowUnknownFields()

	err = dec.Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, err
}

package prometheus

import (
	client "github.com/prometheus/client_golang/api"
	api "github.com/prometheus/client_golang/api/prometheus/v1"
)

func NewAPI(address string) (api.API, error) {
	c, err := client.NewClient(client.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	return api.NewAPI(c), nil
}

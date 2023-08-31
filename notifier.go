package main

import (
	"context"
	"fmt"
	"github.com/rea1shane/prometheus-notifier/prometheus"
	"time"
)

const (
	prometheusAddress = "http://localhost:9090"
	promql            = "up"
)

func main() {
	api, err := prometheus.NewAPI(prometheusAddress)
	if err != nil {
		panic(err)
	}

	query, warnings, err := api.Query(context.Background(), promql, time.Time{})
	if err != nil {
		panic(err)
	}

	fmt.Println(query)
	fmt.Println()
	fmt.Println(warnings)
}

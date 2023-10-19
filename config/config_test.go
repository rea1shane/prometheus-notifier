package config

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	c, err := Load("config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", c)
}

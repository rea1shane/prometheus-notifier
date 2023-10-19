package config

import (
	"github.com/rea1shane/gooooo/data"
	"github.com/rea1shane/gooooo/os"
)

func Load(path string) (c Config, err error) {
	err = os.Load(path, &c, data.YamlFormat)
	return
}

type Config struct {
	Instances     []Instance     `yaml:"instances"`
	Notifications []Notification `yaml:"notifications"`
}

type Instance struct {
	Name          string `yaml:"name"`
	PrometheusURL string `yaml:"prometheus_url"`
	WecomBotKey   string `yaml:"wecom_bot_key"`
}

type Notification struct {
	Name    string `yaml:"name"`
	Expr    string `yaml:"expr"`
	Crontab string `yaml:"crontab"`
	Message string `yaml:"message"`
}

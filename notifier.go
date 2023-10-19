package main

import (
	"context"
	"fmt"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	mycron "github.com/rea1shane/gooooo/cron"
	"github.com/rea1shane/gooooo/log"
	"github.com/rea1shane/prometheus-notifier/config"
	"github.com/rea1shane/prometheus-notifier/prometheus"
	"github.com/rea1shane/prometheus-notifier/wecom"
	"github.com/robfig/cron/v3"
	"strings"
	"text/template"
	"time"
)

const (
	logPath = "config/config.yaml"
)

func main() {
	formatter := log.NewFormatter()
	formatter.FieldsOrder = []string{"module"}
	logger := log.NewLogger()
	logger.SetFormatter(formatter)
	logger.Info("开始运行")

	// 配置文件
	conf, err := config.Load(logPath)
	if err != nil {
		logger.Fatal(err)
	}

	cronLogger := mycron.GenerateLogger(logger, []string{
		"now",
		"next",
	})
	c := cron.New(cron.WithLogger(cronLogger))

	err = schedule(c, cronLogger, conf.Instances, conf.Notifications)
	if err != nil {
		panic(err)
	}
	c.Start()
	time.Sleep(10 * time.Minute)
	c.Stop()
}

type data struct {
	Labels map[string]string
	Value  float64
}

func schedule(c *cron.Cron, cronLogger *mycron.Logger, instances []config.Instance, notifications []config.Notification) error {
	for _, instance := range instances {
		// 初始化 API
		a, err := prometheus.NewAPI(instance.PrometheusURL)
		if err != nil {
			return err
		}
		// 调度
		for _, notification := range notifications {
			id, err := c.AddFunc(notification.Crontab, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				msgs := query(ctx, a, notification)
				content := fmt.Sprintf("%s\n\n\n%s", notification.Name, strings.Join(msgs, "\n"))
				err := wecom.SendBotMarkdownMsg(instance.WecomBotKey, content)
				if err != nil {
					panic(err)
				}
			})
			if err != nil {
				return err
			}
			cronLogger.RegisterEntry(id, fmt.Sprintf("%s > %s", instance.Name, notification.Name))
		}
	}
	return nil
}

func query(ctx context.Context, api v1.API, notification config.Notification) (msgs []string) {
	// 查询
	samples, err := prometheus.Query(ctx, api, notification.Expr, time.Time{})
	if err != nil {
		panic(err)
	}
	for _, sample := range samples {
		message, err := generateMessage(notification.Message, sample)
		if err != nil {
			panic(err)
		}
		msgs = append(msgs, message)
	}
	return
}

// generateMessage 生成消息
func generateMessage(msg string, sample *model.Sample) (string, error) {
	defs := []string{
		"{{$labels := .Labels}}",
		"{{$value := .Value}}",
	}
	parse, err := template.New("message").Parse(strings.Join(append(defs, msg), ""))
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := parse.Execute(&buf, data{
		Labels: convertToMap(sample.Metric),
		Value:  float64(sample.Value),
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// convertToMap 将 model.Metric 转换为 map[string]string
func convertToMap(labels model.Metric) map[string]string {
	m := make(map[string]string)
	for k, v := range labels {
		m[string(k)] = string(v)
	}
	return m
}

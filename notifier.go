package main

import (
	"context"
	"fmt"
	"github.com/morikuni/failure"
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
	// 初始化日志
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

	// 初始化 cron
	cronLogger := mycron.GenerateLogger(logger, []string{
		"now",
		"next",
	})
	c := cron.New(cron.WithLogger(cronLogger))

	// 初始化调度
	err = schedule(c, cronLogger, conf.Instances, conf.Notifications)
	if err != nil {
		logger.Fatal(err)
	}

	// 开始调度
	c.Start()
	defer c.Stop()
	time.Sleep(10 * time.Minute)
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
			return failure.Wrap(
				err,
				failure.Message("初始化 API 失败"),
				failure.Context{
					"Prometheus URL": instance.PrometheusURL,
				},
			)
		}
		// 调度
		for _, notification := range notifications {
			id, err := c.AddFunc(notification.Crontab, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				msgs, warnings, err := query(ctx, a, notification)
				if err != nil {
					cronLogger.Error(err, "查询失败")
					return
				}
				if warnings != nil {
					for _, warning := range warnings {
						cronLogger.Info(warning)
					}
				}
				content := fmt.Sprintf("**%s**\n\n\n%s", notification.Name, strings.Join(msgs, "\n"))
				err = wecom.SendBotMarkdownMsg(instance.WecomBotKey, content)
				if err != nil {
					cronLogger.Error(err, "发送机器人消息失败")
				}
			})
			if err != nil {
				return failure.Wrap(
					err,
					failure.Message("添加调度任务失败"),
					failure.Context{
						"Instance":     instance.Name,
						"Notification": notification.Name,
					},
				)
			}
			cronLogger.RegisterEntry(id, fmt.Sprintf("%s > %s", instance.Name, notification.Name))
		}
	}
	return nil
}

func query(ctx context.Context, api v1.API, notification config.Notification) (msgs []string, warnings v1.Warnings, err error) {
	// 查询
	samples, warnings, err := prometheus.Query(ctx, api, notification.Expr, time.Time{})
	if err != nil {
		return nil, nil, failure.Wrap(
			err,
			failure.Message("查询失败"),
			failure.Context{
				"PromQL": notification.Expr,
			},
		)
	}
	for _, sample := range samples {
		message, err := generateMessage(notification.Message, sample)
		if err != nil {
			return nil, warnings, failure.Wrap(
				err,
				failure.Message("生成消息失败"),
				failure.Context{
					"Message": notification.Message,
					"Sample":  sample.String(),
				},
			)
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

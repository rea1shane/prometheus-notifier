package prometheus

import (
	"context"
	"github.com/morikuni/failure"
	client "github.com/prometheus/client_golang/api"
	api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

func NewAPI(address string) (api.API, error) {
	c, err := client.NewClient(client.Config{
		Address: address,
	})
	if err != nil {
		return nil, failure.Wrap(err,
			failure.Message("创建 API 失败"),
			failure.Context{
				"Address": address,
			},
		)
	}
	return api.NewAPI(c), nil
}

func Query(ctx context.Context, a api.API, promql string, t time.Time) ([]*model.Sample, api.Warnings, error) {
	result, warnings, err := a.Query(ctx, promql, t)
	if err != nil {
		return nil, nil, failure.Wrap(err,
			failure.Message("查询失败"),
			failure.Context{
				"PromQL": promql,
			},
		)
	}

	// 解析结果
	switch i := result.(type) {
	case model.Vector:
		return i, warnings, err
	}

	return nil, nil, failure.New(failure.StringCode("解析查询结果失败"))
}

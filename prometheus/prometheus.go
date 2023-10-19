package prometheus

import (
	"context"
	"fmt"
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
		return nil, err
	}
	return api.NewAPI(c), nil
}

func Query(ctx context.Context, a api.API, promql string, t time.Time) ([]*model.Sample, error) {
	result, warnings, err := a.Query(ctx, promql, t)
	if err != nil {
		return nil, err
	}
	fmt.Println(warnings)

	// 解析结果
	switch i := result.(type) {
	case nil:
		fmt.Println("查询结果为空")
	case model.Vector:
		return i, err
	default:
		fmt.Println(i)
	}

	return nil, nil
}

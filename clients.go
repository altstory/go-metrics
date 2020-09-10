package metrics

import (
	"context"

	"github.com/altstory/go-metrics/internal/client/aliyun"

	"github.com/altstory/go-metrics/internal/client/tencentcloud"
	"github.com/altstory/go-runner"
)

func init() {
	runner.AddClient("metrics.tencentcloud", func(ctx context.Context, config *tencentcloud.Config) error {
		if config == nil {
			return nil
		}

		return tencentcloud.Register(ctx, config)
	})

	runner.AddClient("metrics.aliyun", func(ctx context.Context, config *aliyun.Config) error {
		if config == nil {
			return nil
		}

		return aliyun.Register(ctx, config)
	})
}

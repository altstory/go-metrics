package metrics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/altstory/go-metrics/internal/client"
	"github.com/altstory/go-metrics/internal/client/tencentcloud"
	"github.com/huandu/go-assert"
)

func TestMetricsSendTencentCloud(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	tencentcloud.Register(ctx, &tencentcloud.Config{
		SecretID:  "XXXXXXXXXXXXXXXXXXXXXXXX",
		SecretKey: "XXXXXXXXXXXXXXXXXXXXXXXX",
		Region:    "XXXXXXXXXXXXXXXXXXXXXXXX",
	})
	m := &Metrics{
		client: client.Default(),
	}

	// 测试同时发送超过腾讯云 30 个指标限制的指标。
	const N = 40
	for i := 0; i < N; i++ {
		metric := m.Define(&Def{
			Category: fmt.Sprintf("metric_%v", i+1),
			Method:   Sum,
		})
		metric.AddForTag("/foo/bar", 1)
	}

	a.NilError(m.Send(ctx))

	// 确保所有统计数据已经清空了。
	stats := m.stats(time.Now())
	a.Assert(len(stats.Metrics) == 0)
}

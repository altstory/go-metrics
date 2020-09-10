package metrics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/altstory/go-metrics/internal/client"
	"github.com/altstory/go-metrics/internal/client/aliyun"
	"github.com/huandu/go-assert"
)

func TestMetricsSendAliYun(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	aliyun.Register(ctx, &aliyun.Config{
		AccessKeyID:  "XXXXXXXXXXXXXXXXXXXXXXXX",
		AccessSecret: "XXXXXXXXXXXXXXXXXXXXXXXX",
		Region:       "XXXXXXXXXXXXXXXXXXXXXXXX",
		GroupID:      "XXXXXXXXXXXXXXXXXXXXXXXX",
		Prefix:       "XXXXXXXXXXXXXXXXXXXXXXXX",
	})
	m := &Metrics{
		client: client.Default(),
	}

	// 免费额度只有 10 个，测试的时候不能超过 N+1 个 tag。
	const N = 4
	metric := m.Define(&Def{
		Category: "metric_test",
		Method:   Sum,
	})

	for i := 0; i < N; i++ {
		metric.AddForTag(fmt.Sprintf("/api_%v", i+1), 1)
	}

	a.NilError(m.Send(ctx))

	// 确保所有统计数据已经清空了。
	stats := m.stats(time.Now())
	a.Assert(len(stats.Metrics) == 0)
}

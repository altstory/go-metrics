package metrics

import (
	"testing"
	"time"

	"github.com/altstory/go-metrics/internal/client"
	"github.com/huandu/go-assert"
)

func TestMetrics(t *testing.T) {
	a := assert.New(t)
	m := &Metrics{}
	now := time.Now()

	proctime := m.Define(&Def{
		Category: "proctime",
		Method:   Average,
	})
	maxProctime := m.Define(&Def{
		Category: "max_proctime",
		Method:   Maximum,
	})
	qps := m.Define(&Def{
		Category: "api_qps",
		Method:   Sum,
		Duration: time.Second,
	})
	apiCount := m.Define(&Def{
		Category: "api_count",
		Method:   Sum,
	})

	const N = 100
	const seconds = 13
	const uri = "/foo/bar"
	var sum int64
	data := make([]int64, 0, N)
	for i := 0; i < N; i++ {
		data = append(data, int64(i+1))
	}

	for _, d := range data {
		sum += int64(d)
		proctime.AddForTag(uri, d)
		maxProctime.Add(d)
		qps.AddForTag(uri, d)
		apiCount.AddForTag(uri, d)
	}

	now = now.Add(seconds * time.Second)
	stats := m.stats(now)
	a.Equal(stats, &client.Stats{
		Time: now,
		Metrics: []client.Metric{
			{
				Name:  "proctime",
				Value: sum / N,
			},
			{
				Name:  "proctime",
				Tag:   "/foo/bar",
				Value: sum / N,
			},
			{
				Name:  "max_proctime",
				Value: N,
			},
			{
				Name:  "api_qps",
				Value: sum / seconds,
			},
			{
				Name:  "api_qps",
				Tag:   "/foo/bar",
				Value: sum / seconds,
			},
			{
				Name:  "api_count",
				Value: sum,
			},
			{
				Name:  "api_count",
				Tag:   "/foo/bar",
				Value: sum,
			},
		},
	})
}

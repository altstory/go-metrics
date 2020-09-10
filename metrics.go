package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/altstory/go-log"
	"github.com/altstory/go-metrics/internal/client"
	"github.com/altstory/go-runner"
)

// Metrics 代表一个统计客户端。
type Metrics struct {
	client client.Client

	mu     sync.Mutex
	values []*Value

	flush chan bool
}

var defaultMetrics = &Metrics{
	flush: make(chan bool),
}

func init() {
	runner.OnStart(func(ctx context.Context) error {
		defaultMetrics.client = client.Default()

		if defaultMetrics.client != nil {
			go defaultMetrics.loop()
		}

		return nil
	})

	runner.OnExit(func(ctx context.Context) {
		Flush(ctx)
	})
}

// Define 定一个监控指标。
func Define(def *Def) *Metric {
	return defaultMetrics.Define(def)
}

// Flush 将未发送的统计信息发出去。
func Flush(ctx context.Context) {
	if defaultMetrics.client == nil {
		// 没有启动任何客户端，通过日志输出当前所有的统计信息。
		stats := defaultMetrics.stats(time.Now())

		for _, m := range stats.Metrics {
			log.Tracef(ctx, "name=%v||value=%v||go-metrics: dump metrics", m.Name, m.Value)
		}

		return
	}

	defaultMetrics.flush <- true
}

// Define 定一个监控指标。
func (m *Metrics) Define(def *Def) *Metric {
	now := time.Now()
	v := NewValue(now, def)
	metric := NewMetric(v)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = append(m.values, v)
	return metric
}

// Send 对外发送指标。
func (m *Metrics) Send(ctx context.Context) error {
	if m.client == nil {
		return nil
	}

	now := time.Now()
	stats := m.stats(now)

	if len(stats.Metrics) == 0 {
		return nil
	}

	return m.client.Send(ctx, stats)
}

func (m *Metrics) stats(now time.Time) *client.Stats {
	return &client.Stats{
		Time:    now,
		Metrics: m.metrics(now),
	}
}

func defaultFormatName(category, tag string) string {
	if tag == "" {
		return category
	}

	return fmt.Sprintf("%v:%v", category, tag)
}

func (m *Metrics) metrics(now time.Time) (metrics []client.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range m.values {
		entries := v.Read(now)

		for _, entry := range entries {
			metrics = append(metrics, client.Metric{
				Name:  entry.Category,
				Tag:   entry.Tag,
				Value: entry.Value,
			})
		}
	}

	return
}

func (m *Metrics) loop() {
	var err error
	ctx := context.Background()
	interval := m.client.Interval()
	timer := time.NewTimer(interval)

	for {
		select {
		case <-m.flush:
			if !timer.Stop() {
				<-timer.C
			}

			timer.Reset(interval)
			err = m.Send(ctx)

		case <-timer.C:
			timer.Reset(interval)
			err = m.Send(ctx)
		}

		if err != nil {
			log.Errorf(ctx, "err=%v||go-metrics: fail to send out metrics", err)
			err = nil
		}
	}
}

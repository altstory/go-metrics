package client

import (
	"context"
	"time"
)

// Client 代表一个 metrics 统计客户端的通用接口。
type Client interface {
	// Interval 返回发送的间隔时间。
	Interval() time.Duration

	// 将统计信息同步的发送给平台。
	Send(ctx context.Context, stats *Stats) error
}

// Stats 代表当前需要发送的所有统计信息。
type Stats struct {
	Time    time.Time // 统计发生的时间。
	Metrics []Metric  // 所有统计数据。
}

// Metric 是一个统计维度。
type Metric struct {
	Name  string // 统计项名字。
	Tag   string // 额外标签名，比如统计 api_proctime 时候的 api 名字。
	Value int64  // 统计值。
}

var defaultClient Client

// Default 返回一个已经初始化的 client。
func Default() Client {
	return defaultClient
}

// SetDefault 设置默认 client。
func SetDefault(client Client) {
	defaultClient = client
}

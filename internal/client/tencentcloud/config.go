package tencentcloud

import "time"

const (
	// DefaultEndPoint 是默认的监控上报地址。
	DefaultEndPoint = "monitor.tencentcloudapi.com"

	// DefaultTimeout 是默认的请求超时时间。
	DefaultTimeout = 5 * time.Second

	// DefaultSendInterval 是默认的统计周期，发送请求的时候最多以这样的频率发送统计数据。
	DefaultSendInterval = time.Minute
)

// Config 代表腾讯云 metrics 监控的配置。
type Config struct {
	SecretID     string        `config:"secret_id"`     // 腾讯云帐号的 secret id。
	SecretKey    string        `config:"secret_key"`    // 腾讯云帐号的 secret key。
	Region       string        `config:"region"`        // 腾讯云的大区，比如 ap-guangzhou。
	EndPoint     string        `config:"end_point"`     // 腾讯云的监控上报地址，默认是 DefaultEndPoint。
	Prefix       string        `config:"prefix"`        // 腾讯云统计的前缀。
	Timeout      time.Duration `config:"timeout"`       // 发送统计请求时的最大超时，默认是 DefaultTimeout。
	SendInterval time.Duration `config:"send_interval"` // 发送统计的周期，默认为 DefaultSendInterval。
}

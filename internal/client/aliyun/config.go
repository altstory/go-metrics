package aliyun

import "time"

const (
	// DefaultPeriod 是默认的统计周期，发送请求的时候最多以这样的频率发送统计数据。
	DefaultPeriod = time.Minute
)

// Config 代表阿里云 metrics 监控的配置。
type Config struct {
	AccessKeyID  string        `config:"access_key_id"` // 阿里云帐号的 secret id。
	AccessSecret string        `config:"access_secret"` // 阿里云帐号的 secret key。
	Region       string        `config:"region"`        // 阿里云的大区，比如 cn-hangzhou。
	GroupID      string        `config:"group_id"`      // 阿里云的应用分组 ID。
	Prefix       string        `config:"prefix"`        // 阿里云统计的前缀。
	Period       time.Duration `config:"period"`        // 发送统计的周期，默认为 DefaultSendInterval。
}

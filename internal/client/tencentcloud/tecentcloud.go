package tencentcloud

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"

	"github.com/altstory/go-log"
	"github.com/altstory/go-metrics/internal/client"
)

// TencentCloud 实现了腾讯云 metrics 客户端。
type TencentCloud struct {
	client   *monitor.Client
	prefix   string
	interval time.Duration
}

var _ client.Client = new(TencentCloud)

var (
	defaultClient *TencentCloud
	hostname      string
)

// Register 将腾讯云 metrics 客户端初始化。
func Register(ctx context.Context, config *Config) error {
	name, err := os.Hostname()

	if err != nil {
		log.Errorf(ctx, "err=%v||go-metrics: fail to get hostname", err)
		return err
	}

	ctx = log.WithMoreInfo(ctx,
		log.Info{Key: "hostname", Value: name},
	)
	hostname = name

	if config.SecretID == "" || config.SecretKey == "" || config.Region == "" {
		err = errors.New("go-metrics: missing required settings in config")
		log.Errorf(ctx, "config=%#v||%v", config, err)
		return err
	}

	if config.SendInterval <= 0 {
		config.SendInterval = DefaultSendInterval
	}

	if config.Timeout <= 0 {
		config.Timeout = DefaultTimeout
	}

	if config.EndPoint == "" {
		config.EndPoint = DefaultEndPoint
	}

	credential := common.NewCredential(
		config.SecretID,
		config.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = config.EndPoint
	cpf.HttpProfile.ReqTimeout = int(time.Second.Round(config.Timeout) / time.Second)
	c, err := monitor.NewClient(credential, "ap-guangzhou", cpf)

	if err != nil {
		log.Errorf(ctx, "err=%v||go-metrics: fail to create tencent cloud metrics client", err)
		return err
	}

	defaultClient = &TencentCloud{
		client:   c,
		prefix:   config.Prefix,
		interval: config.SendInterval,
	}
	client.SetDefault(defaultClient)
	return nil
}

var (
	invalidName = regexp.MustCompile(`[^A-Za-z0-9_\-]+`)
)

// formatName 将 category 和 tag 格式化成一个合法的 metric 名字。
func (tc *TencentCloud) formatName(category, tag string) string {
	b := &strings.Builder{}

	if tc.prefix != "" {
		b.WriteString(tc.prefix)
	}

	b.WriteString(invalidName.ReplaceAllLiteralString(category, "_"))

	if tag == "" {
		return b.String()
	}

	b.WriteString("-")
	tag = invalidName.ReplaceAllLiteralString(tag, "_")
	tag = strings.Trim(tag, "_")
	b.WriteString(tag)
	return b.String()
}

// Interval 返回发送统计的时间间隔。
func (tc *TencentCloud) Interval() time.Duration {
	return tc.interval
}

const maxNumMetrics = 30

// Send 将统计项发送给腾讯云。
func (tc *TencentCloud) Send(ctx context.Context, stats *client.Stats) error {
	metrics := stats.Metrics

	for len(metrics) > 0 {
		n := len(metrics)
		request := monitor.NewPutMonitorDataRequest()
		request.AnnounceInstance = &hostname
		request.AnnounceTimestamp = common.Uint64Ptr(uint64(stats.Time.Unix()))
		request.Metrics = make([]*monitor.MetricDatum, 0, maxNumMetrics)

		for i, m := range metrics {
			// 每次只能发送不超过上限的指标数。
			if i >= maxNumMetrics {
				n = i
				break
			}

			metricName := tc.formatName(m.Name, m.Tag)
			request.Metrics = append(request.Metrics, &monitor.MetricDatum{
				MetricName: common.StringPtr(metricName),
				Value:      common.Uint64Ptr(uint64(m.Value)),
			})
		}

		_, err := defaultClient.client.PutMonitorData(request)

		if err != nil {
			log.Errorf(ctx, "err=%v||platform=tencentcloud||request=%v||go-metrics: fail to send metrics to tencent", err, request.ToJsonString())
			return err
		}

		metrics = metrics[n:]
	}

	return nil
}

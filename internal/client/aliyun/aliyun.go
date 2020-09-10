package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/altstory/go-log"
	"github.com/altstory/go-metrics/internal/client"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

// AliYun 实现了阿里云 metrics 客户端。
type AliYun struct {
	client   *cms.Client
	prefix   string
	groupID  string
	interval time.Duration
}

var _ client.Client = new(AliYun)

var (
	defaultClient *AliYun
	hostname      string
)

// Register 将阿里云 metrics 客户端初始化。
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

	if config.AccessKeyID == "" || config.AccessSecret == "" || config.Region == "" || config.GroupID == "" {
		err = errors.New("go-metrics: missing required settings in config")
		log.Errorf(ctx, "config=%#v||%v", config, err)
		return err
	}

	if config.Period <= 0 {
		config.Period = DefaultPeriod
	}

	c, err := cms.NewClientWithAccessKey(config.Region, config.AccessKeyID, config.AccessSecret)

	if err != nil {
		log.Errorf(ctx, "err=%v||go-metrics: fail to create aliyun cloud metrics client", err)
		return err
	}

	// 给前缀加上一个下划线作为区分。
	prefix := config.Prefix

	if prefix != "" {
		prefix += "_"
	}

	defaultClient = &AliYun{
		client:   c,
		prefix:   prefix,
		interval: config.Period,
		groupID:  config.GroupID,
	}
	client.SetDefault(defaultClient)
	return nil
}

// Interval 返回发送统计的时间间隔。
func (al *AliYun) Interval() time.Duration {
	return al.interval
}

const (
	defaultType      = "0"
	valueFormat      = `{"value":%d}`
	dimensionsFormat = `{"tag":%s}`
	dimensionsAll    = `{"tag":"*"}`
)

// Send 将统计项发送给阿里云。
func (al *AliYun) Send(ctx context.Context, stats *client.Stats) error {
	metrics := stats.Metrics

	metricList := make([]cms.PutCustomMetricMetricList, 0, len(metrics))
	timestamp := strconv.FormatInt(stats.Time.UnixNano()/int64(time.Millisecond), 10)
	period := strconv.FormatInt(int64(al.interval/time.Second), 10)

	for _, m := range metrics {
		// 阿里云会自动将 name 里面的非法字符替换成下划线，所以不需要我们额外做什么。
		name := m.Name

		if al.prefix != "" {
			name = al.prefix + name
		}

		// 维度仅统计 tag，不使用 hostname。
		// 这是因为阿里云根据维度收费，hostname 实在太多了，很容易就超出上限。
		var dimensions string

		if m.Tag != "" {
			str, _ := json.Marshal(m.Tag)
			dimensions = fmt.Sprintf(dimensionsFormat, str)
		} else {
			dimensions = dimensionsAll
		}

		metric := cms.PutCustomMetricMetricList{
			Period:     period,
			GroupId:    al.groupID,
			Values:     fmt.Sprintf(valueFormat, m.Value),
			Time:       timestamp,
			MetricName: name,
			Type:       defaultType,
			Dimensions: dimensions,
		}
		metricList = append(metricList, metric)
	}

	request := cms.CreatePutCustomMetricRequest()
	request.Scheme = "https"
	request.MetricList = &metricList
	resp, err := defaultClient.client.PutCustomMetric(request)

	if err != nil {
		log.Errorf(ctx, "err=%v||platform=aliyun||request=%v||go-metrics: fail to send metrics to aliyun", err, request)
		return err
	}

	if resp.Code != "200" {
		err = fmt.Errorf("go-metrics: fail to call PutCustomMetric in aliyun")
		message := strings.Replace(resp.Message, "\n", "", -1)
		log.Errorf(ctx, "err=%v||platform=aliyun||code=%v||message=%v||request=%v||%v",
			err, resp.Code, message, request, err)
		return err
	}

	return nil
}

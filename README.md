# go-metrics：封装云服务平台后端监控接口 #

`go-metrics` 封装了云服务平台的后端监控接口，当前仅支持腾讯云。

## 使用方法 ##

### 定义 `Metric` ###

`go-metrics` 无需手动进行初始化，框架和业务可以在 `go-runner` 里面注册启动回调，通过 `Define` 接口定义监控统计参数，并使用 `Metric` 对象来不断更新数据。

```go
var httpMetrics struct {
    Count, QPS, ProcTime, MaxProcTime *metrics.Metric
}

func init() {
    runner.OnStart(func(ctx context.Context) error {
        // 统计项在监控系统里面指标就是 api_count。
        // 这里统计的数据会累加。
        httpMetrics.Count = metrics.Define(&metrics.Def{
            Category: "api_count",
            Method: metrics.Sum,
        })

        // 这样设置可以计算 QPS。
        httpMetrics.QPS = metrics.Define(&metrics.Def{
            Category: "api_qps",
            Method: metrics.Sum,
            Duration: time.Second,
        })

        // 这样设置可以计算平均处理时间。
        httpMetrics.ProcTime = metrics.Define(&metrics.Def{
            Category: "api_proc_time",
            Method: metrics.Average,
        })

        // 这样设置可以找到最大处理时间。
        httpMetrics.MaxProcTime = metrics.Define(&metrics.Def{
            Category: "api_max_proc_time",
            Method: metrics.Maximum,
        })
    })
}
```

### 使用 `Metric` ###

`Metric` 可以并发调用，这个函数不会阻塞

```go
// 在业务代码里面可以直接使用 metricFoo 来进行业务数据统计。
metricAPICount.Add(1)
```

如果我们想统计更加细分的信息，可以使用 `AddForTag`。这个调用不但会在细分指标里面计数，也会自动调用一次 `Add` 从而在总数里面也加上对应的数据。

```go
// 也可以指定一个 tag 指标，这种用法主要用来统计一个细分指标。
// 比如，我们需要统计 /foo/bar 这个 api 的调用次数，我们可以像下面这样写，
// 这样我们会在 api_count 里面加 1，同时会在 api_count-foo_bar 这个指标项也加 1。
metricAPICount.AddForTag("/foo/bar", 1)
```

## 腾讯云配置 ##

如果指定了 `[metrics.tencentcloud]` 配置，那么就会自动启动腾讯云的客户端。如果不配置，则不会启动这个客户端，数据只会统计但不会做任何的上报。

经典的配置如下。

```ini
[metrics.tencentcloud]
secret_id  = 'XXXXXXXXXXXXXXXXXXXXXXXX'
secret_key = 'XXXXXXXXXXXXXXXXXXXXXXXX'
region = 'ap-guangzhou'
```

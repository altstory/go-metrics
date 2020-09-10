package metrics

// Metric 是一个指标项，用于在业务代码调用，设置指标项。
type Metric struct {
	value *Value
}

// NewMetric 创建一个指标，具体的数据记录在 value 里面。
func NewMetric(value *Value) *Metric {
	return &Metric{
		value: value,
	}
}

// Add 为当前的统计指标增加一定量的数值。
func (m *Metric) Add(value int64) {
	if m == nil || m.value == nil {
		return
	}

	m.value.Add(value)
}

// AddForTag 为指定的 tag 增加一定量的数值。
// tag 一般用于给统计指标增加一些额外信息。
//
// 注意，当调用 AddForTag 时，Add 也会被自动调一次。
func (m *Metric) AddForTag(tag string, value int64) {
	if m == nil || m.value == nil {
		return
	}

	m.value.AddForTag(tag, value)
}

package metrics

import "time"

// Def 代表一个统计字段的定义。
//
// 当需要统计类似 qps 这样的指标时候，应该选择使用 Method = Sum，
// 并且将 Duration = time.Second，这样就能够实现这个功能。
type Def struct {
	Category string        // 统计项的类别。
	Method   Method        // 统计方法。
	Duration time.Duration // 统计周期，如果为 0 代表不按周期进行统计。
}

// Method 代表统计方法。
type Method int

// 各种统计方法。
const (
	Sum Method = iota
	Average
	Maximum
)

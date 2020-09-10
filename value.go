package metrics

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Value 是一类统计值。
// Value 本身不保证并发安全，由调用者自己来保证。
type Value struct {
	def *Def
	ptr unsafe.Pointer
}

type value struct {
	Main counter
	Tags sync.Map

	CreatedAt time.Time
	Updated   int32
}

// Entry 是一个统计项。
type Entry struct {
	Category string
	Tag      string
	Value    int64
}

type counter struct {
	Count int64
	Value int64
}

func (ct *counter) Add(method Method, value int64) {
	if method == Maximum {
		old := atomic.LoadInt64(&ct.Value)

		for value > old {
			old = value
			value = atomic.SwapInt64(&ct.Value, value)
		}
	} else {
		atomic.AddInt64(&ct.Value, value)
	}

	atomic.AddInt64(&ct.Count, 1)
}

// NewValue 创建一个新的 Value 用于统计。
func NewValue(now time.Time, def *Def) *Value {
	v := &Value{
		def: def,
	}
	v.setValue(&value{
		CreatedAt: now,
	})
	return v
}

// Add 增加主要计数器的值。
func (v *Value) Add(value int64) {
	val := v.value()
	val.Main.Add(v.def.Method, value)
	atomic.StoreInt32(&val.Updated, 1)
}

// AddForTag 增加 tag 的计数器的值，同时也会增加主计数器的值。
func (v *Value) AddForTag(tag string, value int64) {
	if tag != "" {
		val := v.value()
		entry, ok := val.Tags.Load(tag)

		if !ok {
			entry = &counter{}
			entry, _ = val.Tags.LoadOrStore(tag, entry)
		}

		counter := entry.(*counter)
		counter.Add(v.def.Method, value)
	}

	v.Add(value)
}

// Read 读取当前最新统计值，并且重置所有的统计。
func (v *Value) Read(now time.Time) (entries []Entry) {
	val := &value{
		CreatedAt: now,
	}
	val = v.setValue(val)

	if atomic.LoadInt32(&val.Updated) == 0 {
		return
	}

	duration := now.Sub(val.CreatedAt)

	// 首先收集所有指标。
	entries = append(entries, Entry{
		Category: v.def.Category,
		Value:    v.calc(&val.Main, duration),
	})
	val.Tags.Range(func(key, value interface{}) bool {
		entries = append(entries, Entry{
			Category: v.def.Category,
			Tag:      key.(string),
			Value:    v.calc(value.(*counter), duration),
		})
		return true
	})

	return
}

func (v *Value) calc(ct *counter, duration time.Duration) (result int64) {
	if ct.Count == 0 {
		return
	}

	switch v.def.Method {
	case Average:
		result = ct.Value / ct.Count
	case Sum, Maximum:
		result = ct.Value
	}

	if v.def.Duration > 0 && v.def.Method != Maximum {
		result = int64(math.Round(float64(result) * v.def.Duration.Seconds() / duration.Seconds()))
	}

	return
}

func (v *Value) value() *value {
	return (*value)(atomic.LoadPointer(&v.ptr))
}

func (v *Value) setValue(val *value) *value {
	old := atomic.SwapPointer(&v.ptr, unsafe.Pointer(val))
	return (*value)(old)
}

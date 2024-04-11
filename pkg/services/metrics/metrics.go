package metrics

import (
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Collector struct {
		bucket     timeBucket
		counter    atomic.Uint64
		cumulative atomic.Int64
	}

	timeBucket struct {
		// More lock free algorithm can be implemented
		// using ring-buffer -> for this kind of application
		// it will be fine... (until it is not)
		durations []time.Duration
		count     int
		full      bool
		mu        sync.Mutex
	}

	Metrics struct {
		Count       uint64
		AverageTime time.Duration
		Cumulate    time.Duration
		P99         time.Duration
		P95         time.Duration
		P90         time.Duration
		P75         time.Duration
		P50         time.Duration
	}
)

func New(size int) *Collector {
	return &Collector{
		bucket: timeBucket{
			durations: make([]time.Duration, size),
			count:     0,
		},
	}
}

func (m *Collector) Collect() Metrics {
	cumulativeTime := m.cumulative.Load()
	count := m.counter.Load()
	bucket := m.bucket.Copy()

	var avg uint64

	for _, d := range bucket.durations {
		avg += uint64(d)
	}

	return Metrics{
		Count:       count,
		AverageTime: time.Duration(avg / count),
		Cumulate:    time.Duration(cumulativeTime),
		P99:         bucket.ExtractP(0.99),
		P95:         bucket.ExtractP(0.95),
		P90:         bucket.ExtractP(0.90),
		P75:         bucket.ExtractP(0.75),
		P50:         bucket.ExtractP(0.50),
	}
}

func (m *Collector) Do(duration time.Duration) {
	m.counter.Add(1)
	m.cumulative.Add(int64(duration))
	m.bucket.Append(duration)
}

func (p *timeBucket) Lock() {
	p.mu.Lock()
}

func (p *timeBucket) Unlock() {
	p.mu.Unlock()
}

func (p *timeBucket) Append(duration time.Duration) {
	p.Lock()
	defer p.Unlock()

	if len(p.durations) == p.count {
		p.count = 0
		p.full = true
	}

	p.durations[p.count] = duration
	p.count++
}

func (p *timeBucket) idx(percentile float64) int {
	idx := p.count

	if p.full {
		idx = len(p.durations)
	}

	// Round and subtract 1
	return int(float64(idx)*percentile+0.5) - 1
}

func (p *timeBucket) ExtractP(percentile float64) time.Duration {
	return p.durations[p.idx(percentile)]
}

func (p *timeBucket) Copy() *timeBucket {
	p.Lock()
	l := p.count

	if p.full {
		l = len(p.durations)
	}

	// Shallow clone is fine -> int64(time.Duration) is value
	durations := slices.Clone(p.durations[:l])
	p.Unlock()
	slices.Sort(durations)

	return &timeBucket{
		durations: durations,
		count:     l,
	}
}

func (p Metrics) String() string {
	var b strings.Builder

	_, _ = b.WriteString("Metrics:")

	_, _ = b.WriteString("\r\nAverage Time: ")
	_, _ = b.WriteString(p.AverageTime.String())

	_, _ = b.WriteString("\r\nCount: ")
	_, _ = b.WriteString(strconv.FormatUint(p.Count, 10))

	_, _ = b.WriteString("\r\nExecution Time: ")
	_, _ = b.WriteString(p.Cumulate.String())

	_, _ = b.WriteString("\r\nP99: ")
	_, _ = b.WriteString(p.P99.String())

	_, _ = b.WriteString("\r\nP95: ")
	_, _ = b.WriteString(p.P95.String())

	_, _ = b.WriteString("\r\nP90: ")
	_, _ = b.WriteString(p.P90.String())

	_, _ = b.WriteString("\r\nP75: ")
	_, _ = b.WriteString(p.P75.String())

	_, _ = b.WriteString("\r\nP50: ")
	_, _ = b.WriteString(p.P50.String())

	return b.String()
}

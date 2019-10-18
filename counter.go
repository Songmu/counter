package counter

import (
	"math"
	"strconv"
	"sync"
	"time"
)

type Counter struct {
	i             uint64
	mu            sync.RWMutex
	max           uint64
	lastRotatedAt time.Time
}

// Count returns the counter
func (c *Counter) Count() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.i
}

// Value is alias method of Count
func (c *Counter) Value() uint64 {
	return c.Count()
}

// String for stringer and expvar.Var
func (c *Counter) String() string {
	return strconv.FormatUint(c.Count(), 10)
}

// Add counter
func (c *Counter) Add(delta uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.max == 0 {
		c.max = math.MaxUint64
	}
	remain := c.max - c.i
	if delta > remain {
		c.lastRotatedAt = time.Now()
		c.i = (delta - remain - 1) % c.max
		return
	}
	c.i += delta
}

// Incr shortcut of Add(1)
func (c *Counter) Incr() {
	c.Add(1)
}

type Observer struct {
	*Counter
	lastObservedAt    time.Time
	lastObservedValue uint64
	mu                sync.Mutex
}

func (o *Observer) Delta() (delta uint64, lastObservedAt, now time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()

	now = time.Now()
	v := o.Count()
	defer func(ti time.Time, v uint64) {
		o.lastObservedAt = now
		o.lastObservedValue = v
	}(now, v)

	if o.lastObservedAt.Before(o.lastRotatedAt) {
		v = o.max - o.lastObservedValue + 1 + v
	} else {
		v = v - o.lastObservedValue
	}
	return v, o.lastObservedAt, now
}

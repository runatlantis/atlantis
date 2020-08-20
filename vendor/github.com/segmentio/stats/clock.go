package stats

import "time"

// The Clock type can be used to report statistics on durations.
//
// Clocks are useful to measure the duration taken by sequential execution steps
// and therefore aren't safe to be used concurrently by multiple goroutines.
type Clock struct {
	name  string
	first time.Time
	last  time.Time
	tags  []Tag
	eng   *Engine
}

// Stamp reports the time difference between now and the last time the method
// was called (or since the clock was created).
//
// The metric produced by this method call will have a "stamp" tag set to name.
func (c *Clock) Stamp(name string) {
	c.StampAt(name, time.Now())
}

// StampAt reports the time difference between now and the last time the method
// was called (or since the clock was created).
//
// The metric produced by this method call will have a "stamp" tag set to name.
func (c *Clock) StampAt(name string, now time.Time) {
	c.observe(name, now.Sub(c.last))
	c.last = now
}

// Stop reports the time difference between now and the time the clock was created at.
//
// The metric produced by this method call will have a "stamp" tag set to
// "total".
func (c *Clock) Stop() {
	c.StopAt(time.Now())
}

// StopAt reports the time difference between now and the time the clock was created at.
//
// The metric produced by this method call will have a "stamp" tag set to
// "total".
func (c *Clock) StopAt(now time.Time) {
	c.observe("total", now.Sub(c.first))
}

func (c *Clock) observe(stamp string, d time.Duration) {
	tags := append(c.tags, Tag{"stamp", stamp})
	c.eng.Observe(c.name, d, tags...)
}

package scheduled

import (
	"runtime"

	tally "github.com/uber-go/tally/v4"
)

type RuntimeStatCollector struct {
	runtimeMetrics runtimeMetrics
}

type runtimeMetrics struct {
	cpuGoroutines tally.Gauge
	cpuCgoCalls   tally.Gauge

	memoryAlloc   tally.Gauge
	memoryTotal   tally.Gauge
	memorySys     tally.Gauge
	memoryLookups tally.Gauge
	memoryMalloc  tally.Gauge
	memoryFrees   tally.Gauge

	memoryHeapAlloc    tally.Gauge
	memoryHeapSys      tally.Gauge
	memoryHeapIdle     tally.Gauge
	memoryHeapInuse    tally.Gauge
	memoryHeapReleased tally.Gauge
	memoryHeapObjects  tally.Gauge

	memoryStackInuse       tally.Gauge
	memoryStackSys         tally.Gauge
	memoryStackMSpanInuse  tally.Gauge
	memoryStackMSpanSys    tally.Gauge
	memoryStackMCacheInuse tally.Gauge
	memoryStackMCacheSys   tally.Gauge

	memoryOtherSys tally.Gauge

	memoryGCSys        tally.Gauge
	memoryGCNext       tally.Gauge
	memoryGCLast       tally.Gauge
	memoryGCPauseTotal tally.Gauge
	memoryGCCount      tally.Gauge
}

func NewRuntimeStats(scope tally.Scope) *RuntimeStatCollector {
	runtimeScope := scope.SubScope("runtime")
	cpuScope := runtimeScope.SubScope("cpu")
	memoryScope := runtimeScope.SubScope("memory")
	heapScope := memoryScope.SubScope("heap")
	stackScope := memoryScope.SubScope("stack")
	gcScope := memoryScope.SubScope("gc")
	runtimeMetrics := runtimeMetrics{
		// cpu
		cpuGoroutines: cpuScope.Gauge("goroutines"),
		cpuCgoCalls:   cpuScope.Gauge("cgo_calls"),
		// memory
		memoryAlloc:   memoryScope.Gauge("alloc"),
		memoryTotal:   memoryScope.Gauge("total"),
		memorySys:     memoryScope.Gauge("sys"),
		memoryLookups: memoryScope.Gauge("lookups"),
		memoryMalloc:  memoryScope.Gauge("malloc"),
		memoryFrees:   memoryScope.Gauge("frees"),
		// heap
		memoryHeapAlloc:    heapScope.Gauge("alloc"),
		memoryHeapSys:      heapScope.Gauge("sys"),
		memoryHeapIdle:     heapScope.Gauge("idle"),
		memoryHeapInuse:    heapScope.Gauge("inuse"),
		memoryHeapReleased: heapScope.Gauge("released"),
		memoryHeapObjects:  heapScope.Gauge("objects"),
		// stack
		memoryStackInuse:       stackScope.Gauge("inuse"),
		memoryStackSys:         stackScope.Gauge("sys"),
		memoryStackMSpanInuse:  stackScope.Gauge("mspan_inuse"),
		memoryStackMSpanSys:    stackScope.Gauge("sys"),
		memoryStackMCacheInuse: stackScope.Gauge("mcache_inuse"),
		memoryStackMCacheSys:   stackScope.Gauge("mcache_sys"),
		memoryOtherSys:         memoryScope.Gauge("othersys"),
		// GC
		memoryGCSys:        gcScope.Gauge("sys"),
		memoryGCNext:       gcScope.Gauge("next"),
		memoryGCLast:       gcScope.Gauge("last"),
		memoryGCPauseTotal: gcScope.Gauge("pause_total"),
		memoryGCCount:      gcScope.Gauge("count"),
	}

	return &RuntimeStatCollector{
		runtimeMetrics: runtimeMetrics,
	}
}

func (r *RuntimeStatCollector) Run() {
	// cpu stats
	r.runtimeMetrics.cpuGoroutines.Update(float64(runtime.NumGoroutine()))
	r.runtimeMetrics.cpuCgoCalls.Update(float64(runtime.NumCgoCall()))

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// general
	r.runtimeMetrics.memoryAlloc.Update(float64(memStats.Alloc))
	r.runtimeMetrics.memoryTotal.Update(float64(memStats.TotalAlloc))
	r.runtimeMetrics.memorySys.Update(float64(memStats.Sys))
	r.runtimeMetrics.memoryLookups.Update(float64(memStats.Lookups))
	r.runtimeMetrics.memoryMalloc.Update(float64(memStats.Mallocs))
	r.runtimeMetrics.memoryFrees.Update(float64(memStats.Frees))

	// heap
	r.runtimeMetrics.memoryHeapAlloc.Update(float64(memStats.HeapAlloc))
	r.runtimeMetrics.memoryHeapSys.Update(float64(memStats.HeapSys))
	r.runtimeMetrics.memoryHeapIdle.Update(float64(memStats.HeapIdle))
	r.runtimeMetrics.memoryHeapInuse.Update(float64(memStats.HeapInuse))
	r.runtimeMetrics.memoryHeapReleased.Update(float64(memStats.HeapReleased))
	r.runtimeMetrics.memoryHeapObjects.Update(float64(memStats.HeapObjects))

	// stack
	r.runtimeMetrics.memoryStackInuse.Update(float64(memStats.StackInuse))
	r.runtimeMetrics.memoryStackSys.Update(float64(memStats.StackSys))
	r.runtimeMetrics.memoryStackMSpanInuse.Update(float64(memStats.MSpanInuse))
	r.runtimeMetrics.memoryStackMSpanSys.Update(float64(memStats.MSpanSys))
	r.runtimeMetrics.memoryStackMCacheInuse.Update(float64(memStats.MCacheInuse))
	r.runtimeMetrics.memoryStackMCacheSys.Update(float64(memStats.MCacheSys))
	r.runtimeMetrics.memoryOtherSys.Update(float64(memStats.OtherSys))

	// GC
	r.runtimeMetrics.memoryGCSys.Update(float64(memStats.GCSys))
	r.runtimeMetrics.memoryGCNext.Update(float64(memStats.NextGC))
	r.runtimeMetrics.memoryGCLast.Update(float64(memStats.LastGC))
	r.runtimeMetrics.memoryGCPauseTotal.Update(float64(memStats.PauseTotalNs))
	r.runtimeMetrics.memoryGCCount.Update(float64(memStats.NumGC))

}

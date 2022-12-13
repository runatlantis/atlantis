package crons

import (
	"context"
	"runtime"

	"github.com/uber-go/tally/v4"
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
	runtimeMetrics := runtimeMetrics{
		// cpu
		cpuGoroutines: runtimeScope.Gauge("cpu.goroutines"),
		cpuCgoCalls:   runtimeScope.Gauge("cpu.cgo_calls"),
		// memory
		memoryAlloc:   runtimeScope.Gauge("memory.alloc"),
		memoryTotal:   runtimeScope.Gauge("memory.total"),
		memorySys:     runtimeScope.Gauge("memory.sys"),
		memoryLookups: runtimeScope.Gauge("memory.lookups"),
		memoryMalloc:  runtimeScope.Gauge("memory.malloc"),
		memoryFrees:   runtimeScope.Gauge("memory.frees"),
		// heap
		memoryHeapAlloc:    runtimeScope.Gauge("memory.heap.alloc"),
		memoryHeapSys:      runtimeScope.Gauge("memory.heap.sys"),
		memoryHeapIdle:     runtimeScope.Gauge("memory.heap.idle"),
		memoryHeapInuse:    runtimeScope.Gauge("memory.heap.inuse"),
		memoryHeapReleased: runtimeScope.Gauge("memory.heap.released"),
		memoryHeapObjects:  runtimeScope.Gauge("memory.heap.objects"),
		// stack
		memoryStackInuse:       runtimeScope.Gauge("memory.stack.inuse"),
		memoryStackSys:         runtimeScope.Gauge("memory.stack.sys"),
		memoryStackMSpanInuse:  runtimeScope.Gauge("memory.stack.mspan_inuse"),
		memoryStackMSpanSys:    runtimeScope.Gauge("memory.stack.sys"),
		memoryStackMCacheInuse: runtimeScope.Gauge("memory.stack.mcache_inuse"),
		memoryStackMCacheSys:   runtimeScope.Gauge("memory.stack.mcache_sys"),
		memoryOtherSys:         runtimeScope.Gauge("memory.othersys"),
		// GC
		memoryGCSys:        runtimeScope.Gauge("memory.gc.sys"),
		memoryGCNext:       runtimeScope.Gauge("memory.gc.next"),
		memoryGCLast:       runtimeScope.Gauge("memory.gc.last"),
		memoryGCPauseTotal: runtimeScope.Gauge("memory.gc.pause_total"),
		memoryGCCount:      runtimeScope.Gauge("memory.gc.count"),
	}

	return &RuntimeStatCollector{
		runtimeMetrics: runtimeMetrics,
	}
}

func (r *RuntimeStatCollector) Run(ctx context.Context) error {
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

	return nil
}

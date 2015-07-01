package gonr

import (
	"runtime"
	"time"

	"github.com/neocortical/newrelic"
	"github.com/neocortical/nrmetrics"
	metrics "github.com/rcrowley/go-metrics"
)

var memStats runtime.MemStats

func addMemoryMetrics(p *newrelic.Plugin) {
	allocInUse := metrics.NewGauge()
	allocTotal := metrics.NewGauge()

	pointerLookups := metrics.NewMeter()
	mallocs := metrics.NewMeter()
	frees := metrics.NewMeter()

	allocHeap := metrics.NewGauge()
	heapSys := metrics.NewGauge()
	heapIdle := metrics.NewGauge()
	heapInUse := metrics.NewGauge()
	heapReleased := metrics.NewGauge()
	heapObjects := metrics.NewGauge()

	stackInUse := metrics.NewGauge()
	stackSys := metrics.NewGauge()

	mspanInUse := metrics.NewGauge()
	mspanSys := metrics.NewGauge()
	mcacheInUse := metrics.NewGauge()
	mcacheSys := metrics.NewGauge()
	buckHashSys := metrics.NewGauge()
	gcSys := metrics.NewGauge()
	otherSys := metrics.NewGauge()

	mr := &memReader{
		sampleRate:     time.Minute,
		allocInUse:     allocInUse,
		allocTotal:     allocTotal,
		pointerLookups: pointerLookups,
		mallocs:        mallocs,
		frees:          frees,
		allocHeap:      allocHeap,
		heapSys:        heapSys,
		heapIdle:       heapIdle,
		heapInUse:      heapInUse,
		heapReleased:   heapReleased,
		heapObjects:    heapObjects,
		stackInUse:     stackInUse,
		stackSys:       stackSys,
		mspanInUse:     mspanInUse,
		mspanSys:       mspanSys,
		mcacheInUse:    mcacheInUse,
		mcacheSys:      mcacheSys,
		buckHashSys:    buckHashSys,
		gcSys:          gcSys,
		otherSys:       otherSys,
	}
	mr.Run()

	nrmetrics.AddGaugeMetric(p, allocInUse, nrmetrics.MetricConfig{Name: "Memory/Summary/In Use", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, allocTotal, nrmetrics.MetricConfig{Name: "Memory/Summary/Total Allocated", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, allocHeap, nrmetrics.MetricConfig{Name: "Memory/Heap/Allocated", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, heapSys, nrmetrics.MetricConfig{Name: "Memory/Heap/System", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, heapIdle, nrmetrics.MetricConfig{Name: "Memory/Heap/Idle", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, heapInUse, nrmetrics.MetricConfig{Name: "Memory/Heap/In Use", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, heapReleased, nrmetrics.MetricConfig{Name: "Memory/Heap/Released", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, heapObjects, nrmetrics.MetricConfig{Name: "Memory/Heap/Objects", Unit: "objects", Value: true})
	nrmetrics.AddGaugeMetric(p, stackInUse, nrmetrics.MetricConfig{Name: "Memory/Stack/In Use", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, stackSys, nrmetrics.MetricConfig{Name: "Memory/Stack/System", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, mspanInUse, nrmetrics.MetricConfig{Name: "Memory/MSpan/In Use", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, mspanSys, nrmetrics.MetricConfig{Name: "Memory/MSpan/System", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, mcacheInUse, nrmetrics.MetricConfig{Name: "Memory/MCache/In Use", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, mcacheSys, nrmetrics.MetricConfig{Name: "Memory/MCache/System", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, buckHashSys, nrmetrics.MetricConfig{Name: "Memory/Misc/Buck Hash", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, gcSys, nrmetrics.MetricConfig{Name: "Memory/Misc/GC", Unit: "bytes", Value: true})
	nrmetrics.AddGaugeMetric(p, otherSys, nrmetrics.MetricConfig{Name: "Memory/Misc/Other", Unit: "bytes", Value: true})

	nrmetrics.AddMeterMetric(p, pointerLookups, nrmetrics.MetricConfig{
		Name:   "Memory/Events/Pointer Lookups",
		Unit:   "lookups",
		Rate1:  true,
		Rate5:  true,
		Rate15: true,
	})
	nrmetrics.AddMeterMetric(p, mallocs, nrmetrics.MetricConfig{
		Name:   "Memory/Events/Mallocs",
		Unit:   "mallocs",
		Rate1:  true,
		Rate5:  true,
		Rate15: true,
	})
	nrmetrics.AddMeterMetric(p, frees, nrmetrics.MetricConfig{
		Name:   "Memory/Events/Frees",
		Unit:   "frees",
		Rate1:  true,
		Rate5:  true,
		Rate15: true,
	})
}

type memReader struct {
	sampleRate time.Duration
	allocInUse metrics.Gauge
	allocTotal metrics.Gauge

	lastPointerLookups uint64
	pointerLookups     metrics.Meter
	lastMallocs        uint64
	mallocs            metrics.Meter
	lastFrees          uint64
	frees              metrics.Meter

	allocHeap    metrics.Gauge
	heapSys      metrics.Gauge
	heapIdle     metrics.Gauge
	heapInUse    metrics.Gauge
	heapReleased metrics.Gauge
	heapObjects  metrics.Gauge

	stackInUse metrics.Gauge
	stackSys   metrics.Gauge

	mspanInUse  metrics.Gauge
	mspanSys    metrics.Gauge
	mcacheInUse metrics.Gauge
	mcacheSys   metrics.Gauge
	buckHashSys metrics.Gauge
	gcSys       metrics.Gauge
	otherSys    metrics.Gauge
}

func (mr *memReader) Run() {
	go mr.run()
}

func (mr *memReader) run() {
	ticks := time.Tick(mr.sampleRate)
	for _ = range ticks {
		mr.updateMetrics()
	}
}

func (mr *memReader) updateMetrics() {
	runtime.ReadMemStats(&memStats)

	mr.allocInUse.Update(int64(memStats.Alloc))
	mr.allocTotal.Update(int64(memStats.TotalAlloc))

	mr.pointerLookups.Mark(int64(memStats.Lookups - mr.lastPointerLookups))
	mr.lastPointerLookups = memStats.Lookups
	mr.mallocs.Mark(int64(memStats.Mallocs - mr.lastMallocs))
	mr.lastMallocs = memStats.Mallocs
	mr.frees.Mark(int64(memStats.Frees - mr.lastFrees))
	mr.lastFrees = memStats.Frees

	mr.allocHeap.Update(int64(memStats.HeapAlloc))
	mr.heapSys.Update(int64(memStats.HeapSys))
	mr.heapIdle.Update(int64(memStats.HeapIdle))
	mr.heapInUse.Update(int64(memStats.HeapInuse))
	mr.heapReleased.Update(int64(memStats.HeapReleased))
	mr.heapObjects.Update(int64(memStats.HeapObjects))

	mr.stackInUse.Update(int64(memStats.StackInuse))
	mr.stackSys.Update(int64(memStats.StackSys))

	mr.mspanInUse.Update(int64(memStats.MSpanInuse))
	mr.mspanSys.Update(int64(memStats.MSpanSys))
	mr.mcacheInUse.Update(int64(memStats.MCacheInuse))
	mr.mcacheSys.Update(int64(memStats.MCacheSys))
	mr.buckHashSys.Update(int64(memStats.BuckHashSys))
	mr.gcSys.Update(int64(memStats.GCSys))
	mr.otherSys.Update(int64(memStats.OtherSys))
}

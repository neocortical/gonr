package gonr

import (
	"runtime/debug"
	"time"

	"github.com/neocortical/newrelic"
	"github.com/neocortical/nrmetrics"
	metrics "github.com/rcrowley/go-metrics"
)

var gcStats debug.GCStats

func addGCMetrics(p *newrelic.Plugin) {
	numGC := metrics.NewMeter()
	pauseDur := metrics.NewMeter()
	pauseTime := metrics.NewHistogram(metrics.NewExpDecaySample(10000, 0.015))
	gcr := &gcReader{
		sampleRate: time.Second * 10,
		numGC:      numGC,
		pauseDur:   pauseDur,
		pauseTime:  pauseTime,
	}
	gcr.Run()

	nrmetrics.AddMeterMetric(p, numGC, nrmetrics.MetricConfig{Name: "GC/GC Rate", Unit: "pauses", Rate1: true, Rate5: true, Rate15: true})
	nrmetrics.AddMeterMetric(p, pauseDur, nrmetrics.MetricConfig{Name: "GC/GC Pause Rate", Unit: "nanoseconds", Rate1: true, Rate5: true, Rate15: true})
	nrmetrics.AddHistogramMetric(p, pauseTime, nrmetrics.MetricConfig{
		Name:        "GC/GC Pause Time",
		Unit:        "pauses",
		Duration:    time.Microsecond,
		Mean:        true,
		Percentiles: []float64{0.5, 0.75, 0.9, 0.99, 0.999},
	})
}

type gcReader struct {
	sampleRate     time.Duration
	lastGC         time.Time
	lastNumGC      int64
	numGC          metrics.Meter
	lastPauseTotal time.Duration
	pauseDur       metrics.Meter
	pauseTime      metrics.Histogram
}

func (gcr *gcReader) Run() {
	go gcr.run()
}

func (gcr *gcReader) run() {
	ticks := time.Tick(gcr.sampleRate)
	for _ = range ticks {
		gcr.updateMetrics()
	}
}

func (gcr *gcReader) updateMetrics() {
	debug.ReadGCStats(&gcStats)

	gcr.numGC.Mark(gcStats.NumGC - gcr.lastNumGC)
	gcr.lastNumGC = gcStats.NumGC

	gcr.pauseDur.Mark(int64(gcStats.PauseTotal - gcr.lastPauseTotal))
	gcr.lastPauseTotal = gcStats.PauseTotal

	if gcr.lastGC != gcStats.LastGC && 0 < len(gcStats.Pause) {
		gcr.pauseTime.Update(int64(gcStats.Pause[0]))
	}

	gcr.lastGC = gcStats.LastGC
}

package gonr

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/neocortical/newrelic"
	"github.com/neocortical/nrmetrics"
	metrics "github.com/rcrowley/go-metrics"
)

func addRuntimeMetrics(p *newrelic.Plugin) {
	p.AddMetric(newrelic.NewMetric("Runtime/Goroutines", "goroutines", func() (float64, error) { return float64(runtime.NumGoroutine()), nil }))

	var lastNumCgoCall int64
	p.AddMetric(newrelic.NewMetric("Runtime/CGO Calls", "calls", func() (float64, error) {
		currentNumCgoCall := runtime.NumCgoCall()
		result := float64(currentNumCgoCall - lastNumCgoCall)
		lastNumCgoCall = currentNumCgoCall
		return result, nil
	}))

	threads := metrics.NewGauge()
	fdsize := metrics.NewGauge()
	vmpeak := metrics.NewGauge()
	vmsize := metrics.NewGauge()
	rsspeak := metrics.NewGauge()
	rsssize := metrics.NewGauge()
	pr := &procReader{
		sampleRate: time.Second * 5,
		procMap:    make(map[string]string),
		threads:    threads,
		fdsize:     fdsize,
		vmpeak:     vmpeak,
		vmsize:     vmsize,
		rsspeak:    rsspeak,
		rsssize:    rsssize,
	}
	pr.Run()

	nrmetrics.AddGaugeMetric(p, threads, nrmetrics.MetricConfig{Name: "Runtime/Threads", Unit: "threads", Count: true})
	nrmetrics.AddGaugeMetric(p, fdsize, nrmetrics.MetricConfig{Name: "Runtime/FD Size", Unit: "file descriptor slots", Count: true})
	nrmetrics.AddGaugeMetric(p, vmpeak, nrmetrics.MetricConfig{Name: "Runtime/Peak Virt Mem", Unit: "bytes", Count: true})
	nrmetrics.AddGaugeMetric(p, vmsize, nrmetrics.MetricConfig{Name: "Runtime/Virt Mem", Unit: "bytes", Count: true})
	nrmetrics.AddGaugeMetric(p, rsspeak, nrmetrics.MetricConfig{Name: "Runtime/RSS Peak", Unit: "bytes", Count: true})
	nrmetrics.AddGaugeMetric(p, rsssize, nrmetrics.MetricConfig{Name: "Runtime/RSS Size", Unit: "bytes", Count: true})
}

type procReader struct {
	sampleRate time.Duration
	procMap    map[string]string
	threads    metrics.Gauge
	fdsize     metrics.Gauge
	vmpeak     metrics.Gauge
	vmsize     metrics.Gauge
	rsspeak    metrics.Gauge
	rsssize    metrics.Gauge
}

func (pr *procReader) Run() {
	go pr.run()
}

func (pr *procReader) run() {
	ticks := time.Tick(pr.sampleRate)
	for _ = range ticks {
		pr.updateProcMap()
		pr.updateMetrics()
	}
}

func (pr *procReader) updateProcMap() {
	path := fmt.Sprintf("/proc/%d/status", os.Getpid())
	rawStatus, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(rawStatus), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])

			pr.procMap[k] = v
		}
	}
}

func (pr *procReader) updateMetrics() {
	if val, ok := pr.procMap["Threads"]; ok {
		if threads, err := strconv.ParseInt(val, 10, 64); err == nil {
			pr.threads.Update(threads)
		}
	}
	if val, ok := pr.procMap["FDSize"]; ok {
		if fdsize, err := strconv.ParseInt(val, 10, 64); err == nil {
			pr.fdsize.Update(fdsize)
		}
	}
	if val, ok := pr.procMap["VmPeak"]; ok {
		pr.vmpeak.Update(parseMemValue(val))
	}
	if val, ok := pr.procMap["VmSize"]; ok {
		pr.vmpeak.Update(parseMemValue(val))
	}
	if val, ok := pr.procMap["VmHWM"]; ok {
		pr.vmpeak.Update(parseMemValue(val))
	}
	if val, ok := pr.procMap["VmRSS"]; ok {
		pr.vmpeak.Update(parseMemValue(val))
	}
}

func parseMemValue(val string) int64 {
	parts := strings.Split(val, " ")
	if len(parts) != 2 {
		return 0
	}
	result, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}
	switch parts[1] {
	case "kB":
		result *= 1024
	case "mB":
		result *= 1048576
	case "gB":
		result *= 1073741824
	}
	return result
}

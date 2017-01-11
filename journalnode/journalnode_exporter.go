package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

const (
	namespace = "journalnode"
)

var (
	listenAddress     = flag.String("web.listen-address", ":9070", "Address on which to expose metrics and web interface.")
	metricsPath       = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	journalnodeJmxUrl = flag.String("journalnode.jmx.url", "http://localhost:50070/jmx", "Hadoop JMX URL.")
)

type Exporter struct {
	url                      string
	pnGcCount                prometheus.Gauge
	pnGcTime                 prometheus.Gauge
	cmsGcCount               prometheus.Gauge
	cmsGcTime                prometheus.Gauge
	heapMemoryUsageCommitted prometheus.Gauge
	heapMemoryUsageInit      prometheus.Gauge
	heapMemoryUsageMax       prometheus.Gauge
	heapMemoryUsageUsed      prometheus.Gauge
}

func NewExporter(url string) *Exporter {
	return &Exporter{
		url: url,
		pnGcCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ParNew_CollectionCount",
			Help:      "ParNew GC Count",
		}),
		pnGcTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ParNew_CollectionTime",
			Help:      "ParNew GC Time",
		}),
		cmsGcCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ConcurrentMarkSweep_CollectionCount",
			Help:      "ConcurrentMarkSweep GC Count",
		}),
		cmsGcTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ConcurrentMarkSweep_CollectionTime",
			Help:      "ConcurrentMarkSweep GC Time",
		}),
		heapMemoryUsageCommitted: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageCommitted",
			Help:      "heapMemoryUsageCommitted",
		}),
		heapMemoryUsageInit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageInit",
			Help:      "heapMemoryUsageInit",
		}),
		heapMemoryUsageMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageMax",
			Help:      "heapMemoryUsageMax",
		}),
		heapMemoryUsageUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "heapMemoryUsageUsed",
			Help:      "heapMemoryUsageUsed",
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.pnGcCount.Describe(ch)
	e.pnGcTime.Describe(ch)
	e.cmsGcCount.Describe(ch)
	e.cmsGcTime.Describe(ch)
	e.heapMemoryUsageCommitted.Describe(ch)
	e.heapMemoryUsageInit.Describe(ch)
	e.heapMemoryUsageMax.Describe(ch)
	e.heapMemoryUsageUsed.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	resp, err := http.Get(e.url)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	var f interface{}
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Error(err)
	}
	m := f.(map[string]interface{})
	var journalList = m["beans"].([]interface{})
	for _, journalData := range journalList {
		journalDataMap := journalData.(map[string]interface{})

		if journalDataMap["name"] == "java.lang:type=GarbageCollector,name=ParNew" {
			e.pnGcCount.Set(journalDataMap["CollectionCount"].(float64))
			e.pnGcTime.Set(journalDataMap["CollectionTime"].(float64))
		}
		if journalDataMap["name"] == "java.lang:type=GarbageCollector,name=ConcurrentMarkSweep" {
			e.cmsGcCount.Set(journalDataMap["CollectionCount"].(float64))
			e.cmsGcTime.Set(journalDataMap["CollectionTime"].(float64))
		}
		/*
			"name" : "java.lang:type=Memory",
			"modelerType" : "sun.management.MemoryImpl",
			"HeapMemoryUsage" : {
				"committed" : 1060372480,
				"init" : 1073741824,
				"max" : 1060372480,
				"used" : 124571464
			},
		*/
		if journalDataMap["name"] == "java.lang:type=Memory" {
			heapMemoryUsage := journalDataMap["HeapMemoryUsage"].(map[string]interface{})
			e.heapMemoryUsageCommitted.Set(heapMemoryUsage["committed"].(float64))
			e.heapMemoryUsageInit.Set(heapMemoryUsage["init"].(float64))
			e.heapMemoryUsageMax.Set(heapMemoryUsage["max"].(float64))
			e.heapMemoryUsageUsed.Set(heapMemoryUsage["used"].(float64))
		}
	}
	e.pnGcCount.Collect(ch)
	e.pnGcTime.Collect(ch)
	e.cmsGcCount.Collect(ch)
	e.cmsGcTime.Collect(ch)
	e.heapMemoryUsageCommitted.Collect(ch)
	e.heapMemoryUsageInit.Collect(ch)
	e.heapMemoryUsageMax.Collect(ch)
	e.heapMemoryUsageUsed.Collect(ch)
}

func main() {
	flag.Parse()

	exporter := NewExporter(*journalnodeJmxUrl)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>JournalNode Exporter</title></head>
		<body>
		<h1>JournalNode Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

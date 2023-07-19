package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/meoww-bot/hadoop_exporter/lib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/log"
)

const (
	namespace = "hdfs_namenode"
)

var (
	listenAddress  = flag.String("web.listen-address", ":9070", "Address on which to expose metrics and web interface.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	namenodeJmxUrl = flag.String("namenode.jmx.url", "http://nn01.example.com:50070/jmx", "Hadoop JMX URL.")
	keytabPath     = flag.String("krb5.keytab.path", "", "Kerberos keytab file path")
	principal      = flag.String("krb5.principal", "", "Principal (admin@EXAMPLE.COM)")
)

type Exporter struct {
	url                   string
	keytabPath            string
	principal             string
	MissingBlocks         prometheus.Gauge
	UnderReplicatedBlocks prometheus.Gauge
	Capacity              *prometheus.GaugeVec
	BlocksTotal           prometheus.Gauge
	FilesTotal            prometheus.Gauge
	CorruptBlocks         prometheus.Gauge
	ExcessBlocks          prometheus.Gauge
	StaleDataNodes        prometheus.Gauge
	GcCount               *prometheus.GaugeVec
	GcTime                *prometheus.GaugeVec
	heapMemoryUsage       *prometheus.GaugeVec
	lastHATransitionTime  prometheus.Gauge
	HAState               prometheus.Gauge
	RpcReceivedBytes      *prometheus.GaugeVec
	RpcSentBytes          *prometheus.GaugeVec
	RpcQueueTimeNumOps    *prometheus.GaugeVec // RpcProcessingTimeNumOps = RpcQueueTimeNumOps
	RpcAvgTime            *prometheus.GaugeVec
	RpcNumOpenConnections *prometheus.GaugeVec // current number of open connections
	RpcCallQueueLength    *prometheus.GaugeVec
}

func NewExporter(url string, keytabPath string, principal string) *Exporter {

	return &Exporter{
		url:        url,
		keytabPath: keytabPath,
		principal:  principal,
		MissingBlocks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "missing_blocks",
			Help:      "Current number of missing blocks",
		}),
		UnderReplicatedBlocks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "under_replicated_blocks",
			Help:      "Current number of blocks under replicated",
		}),
		Capacity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "capacity_bytes",
			Help:      "Current DataNodes capacity in each mode in bytes",
		}, []string{"mode"}),
		BlocksTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "blocks_total",
			Help:      "Current number of allocated blocks in the system",
		}),
		FilesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "files_total",
			Help:      "Current number of files and directories",
		}),
		CorruptBlocks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "corrupt_blocks",
			Help:      "Current number of blocks with corrupt replicas",
		}),
		ExcessBlocks: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "excess_blocks",
			Help:      "Current number of excess blocks",
		}),
		StaleDataNodes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "stale_datanodes",
			Help:      "Current number of DataNodes marked stale due to delayed heartbeat",
		}),
		GcCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "jvm_metrics",
			Name:      "gc_count",
			Help:      "GC count of each type",
		}, []string{"type"}),
		GcTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "jvm_metrics",
			Name:      "gc_time_milliseconds",
			Help:      "GC time of each type in milliseconds",
		}, []string{"type"}),
		heapMemoryUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "memory",
			Name:      "heap_memory_usage_bytes",
			Help:      "Current heap memory of each mode in bytes",
		}, []string{"mode"}),
		lastHATransitionTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "namenode_status",
			Name:      "last_ha_transition_time",
			Help:      "last HA Transition Time",
		}),
		HAState: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "fsname_system",
			Name:      "hastate",
			Help:      "Current state of the NameNode: 0.0 (for initializing) or 1.0 (for active) or 2.0 (for standby) or 3.0 (for stopping) state",
		}),
		// RpcActivityForPort8020
		// RpcActivityForPort8060
		RpcReceivedBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "received_bytes",
			Help:      "Total number of received bytes",
		}, []string{"port"}),
		RpcSentBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "sent_bytes",
			Help:      "Total number of sent bytes",
		}, []string{"port"}),
		RpcQueueTimeNumOps: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "call_count",
			Help:      "Total number of RPC calls (same to RpcQueueTimeNumOps) ",
		}, []string{"port", "method"}),
		RpcAvgTime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "avg_time_milliseconds",
			Help:      "current number of open connections",
		}, []string{"port", "method"}),
		RpcNumOpenConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "open_connections_count",
			Help:      "current number of open connections",
		}, []string{"port"}),
		RpcCallQueueLength: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "rpc_activity",
			Name:      "call_queue_length",
			Help:      "Current length of the call queue",
		}, []string{"port"}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.MissingBlocks.Describe(ch)
	e.UnderReplicatedBlocks.Describe(ch)
	e.Capacity.Describe(ch)
	e.BlocksTotal.Describe(ch)
	e.FilesTotal.Describe(ch)
	e.CorruptBlocks.Describe(ch)
	e.ExcessBlocks.Describe(ch)
	e.StaleDataNodes.Describe(ch)
	e.GcCount.Describe(ch)
	e.GcTime.Describe(ch)
	e.heapMemoryUsage.Describe(ch)
	e.lastHATransitionTime.Describe(ch)
	e.HAState.Describe(ch)
	e.RpcReceivedBytes.Describe(ch)
	e.RpcSentBytes.Describe(ch)
	e.RpcQueueTimeNumOps.Describe(ch)
	e.RpcAvgTime.Describe(ch)
	e.RpcNumOpenConnections.Describe(ch)
	e.RpcCallQueueLength.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	var data []byte
	var err error

	if e.keytabPath != "" {
		data = lib.MakeKrb5RequestWithKeytab(e.keytabPath, e.principal, e.url)

	} else {

		resp, err := http.Get(e.url)
		if err != nil {
			log.Error(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			log.Error("HTTP status 401 Unauthorized")
			os.Exit(-1)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
		}
	}
	var f interface{}
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Error(err)
		fmt.Println(data)
	}
	// {"beans":[{"name":"Hadoop:service=NameNode,name=FSNamesystem", ...}, {"name":"java.lang:type=MemoryPool,name=Code Cache", ...}, ...]}
	m := f.(map[string]interface{})
	// [{"name":"Hadoop:service=NameNode,name=FSNamesystem", ...}, {"name":"java.lang:type=MemoryPool,name=Code Cache", ...}, ...]
	var nameList = m["beans"].([]interface{})
	for _, nameData := range nameList {
		nameDataMap := nameData.(map[string]interface{})
		/*
		   {
		       "name" : "Hadoop:service=NameNode,name=FSNamesystem",
		       "modelerType" : "FSNamesystem",
		       "tag.Context" : "dfs",
		       "tag.HAState" : "active",
		       "tag.TotalSyncTimes" : "23 6 ",
		       "tag.Hostname" : "CNHORTO7502.line.ism",
		       "MissingBlocks" : 0,
		       "MissingReplOneBlocks" : 0,
		       "ExpiredHeartbeats" : 0,
		       "TransactionsSinceLastCheckpoint" : 2007,
		       "TransactionsSinceLastLogRoll" : 7,
		       "LastWrittenTransactionId" : 172706,
		       "LastCheckpointTime" : 1456089173101,
		       "CapacityTotal" : 307099828224,
		       "CapacityTotalGB" : 286.0,
		       "CapacityUsed" : 1471291392,
		       "CapacityUsedGB" : 1.0,
		       "CapacityRemaining" : 279994568704,
		       "CapacityRemainingGB" : 261.0,
		       "CapacityUsedNonDFS" : 25633968128,
		       "TotalLoad" : 6,
		       "SnapshottableDirectories" : 0,
		       "Snapshots" : 0,
		       "LockQueueLength" : 0,
		       "BlocksTotal" : 67,
		       "NumFilesUnderConstruction" : 0,
		       "NumActiveClients" : 0,
		       "FilesTotal" : 184,
		       "PendingReplicationBlocks" : 0,
		       "UnderReplicatedBlocks" : 0,
		       "CorruptBlocks" : 0,
		       "ScheduledReplicationBlocks" : 0,
		       "PendingDeletionBlocks" : 0,
		       "ExcessBlocks" : 0,
		       "PostponedMisreplicatedBlocks" : 0,
		       "PendingDataNodeMessageCount" : 0,
		       "MillisSinceLastLoadedEdits" : 0,
		       "BlockCapacity" : 2097152,
		       "StaleDataNodes" : 0,
		       "TotalFiles" : 184,
		       "TotalSyncCount" : 7
		   }
		*/
		if nameDataMap["name"] == "Hadoop:service=NameNode,name=FSNamesystem" {
			e.MissingBlocks.Set(nameDataMap["MissingBlocks"].(float64))
			e.UnderReplicatedBlocks.Set(nameDataMap["UnderReplicatedBlocks"].(float64))
			e.Capacity.WithLabelValues("Total").Set(nameDataMap["CapacityTotal"].(float64))
			e.Capacity.WithLabelValues("Used").Set(nameDataMap["CapacityUsed"].(float64))
			e.Capacity.WithLabelValues("Remaining").Set(nameDataMap["CapacityRemaining"].(float64))
			e.Capacity.WithLabelValues("UsedNonDFS").Set(nameDataMap["CapacityUsedNonDFS"].(float64))
			e.BlocksTotal.Set(nameDataMap["BlocksTotal"].(float64))
			e.FilesTotal.Set(nameDataMap["FilesTotal"].(float64))
			e.CorruptBlocks.Set(nameDataMap["CorruptBlocks"].(float64))
			e.ExcessBlocks.Set(nameDataMap["ExcessBlocks"].(float64))
			e.StaleDataNodes.Set(nameDataMap["StaleDataNodes"].(float64))

			switch nameDataMap["tag.HAState"] {

			case "initializing":
				e.HAState.Set(0)
			case "active":
				e.HAState.Set(1)
			case "standby":
				e.HAState.Set(2)
			case "stopping":
				e.HAState.Set(3)

			}
		}
		/*
		   {
		       "name" : "Hadoop:service=NameNode,name=NameNodeStatus",
		       "modelerType" : "org.apache.hadoop.hdfs.server.namenode.NameNode",
		       "SecurityEnabled" : false,
		       "NNRole" : "NameNode",
		       "HostAndPort" : "namenode1.hdfs.tamr:50071",
		       "LastHATransitionTime" : 1484149009998,
		       "State" : "active"
		   }
		*/
		if nameDataMap["name"] == "Hadoop:service=NameNode,name=NameNodeStatus" {

			e.lastHATransitionTime.Set(nameDataMap["LastHATransitionTime"].(float64))
		}
		/*
			{
				"name": "Hadoop:service=NameNode,name=JvmMetrics",
				"modelerType": "JvmMetrics",
				"tag.Context": "jvm",
				"tag.ProcessName": "NameNode",
				"tag.SessionId": null,
				"tag.Hostname": "osb002.example.com",
				"MemNonHeapUsedM": 127.088585,
				"MemNonHeapCommittedM": 129.57031,
				"MemNonHeapMaxM": -1.0,
				"MemHeapUsedM": 94972.38,
				"MemHeapCommittedM": 152780.81,
				"MemHeapMaxM": 152780.81,
				"MemMaxM": 152780.81,
				"GcCountParNew": 54800,
				"GcTimeMillisParNew": 20067913,
				"GcCountConcurrentMarkSweep": 13,
				"GcTimeMillisConcurrentMarkSweep": 8184,
				"GcCount": 54813,
				"GcTimeMillis": 20076097,
				"GcNumWarnThresholdExceeded": 1,
				"GcNumInfoThresholdExceeded": 5,
				"GcTotalExtraSleepTime": 8912336,
				"ThreadsNew": 0,
				"ThreadsRunnable": 8,
				"ThreadsBlocked": 0,
				"ThreadsWaiting": 13,
				"ThreadsTimedWaiting": 936,
				"ThreadsTerminated": 0,
				"LogFatal": 0,
				"LogError": 80332,
				"LogWarn": 40327688,
				"LogInfo": 1207922583
			}
		*/
		if nameDataMap["name"] == "Hadoop:service=NameNode,name=JvmMetrics" {
			e.GcCount.WithLabelValues("ParNew").Set(nameDataMap["GcCountParNew"].(float64))
			e.GcCount.WithLabelValues("ConcurrentMarkSweep").Set(nameDataMap["GcCountConcurrentMarkSweep"].(float64))

			e.GcTime.WithLabelValues("ParNew").Set(nameDataMap["GcTimeMillisParNew"].(float64))
			e.GcTime.WithLabelValues("ConcurrentMarkSweep").Set(nameDataMap["GcTimeMillisConcurrentMarkSweep"].(float64))

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
		if nameDataMap["name"] == "java.lang:type=Memory" {
			heapMemoryUsage := nameDataMap["HeapMemoryUsage"].(map[string]interface{})
			e.heapMemoryUsage.WithLabelValues("committed").Set(heapMemoryUsage["committed"].(float64))
			e.heapMemoryUsage.WithLabelValues("init").Set(heapMemoryUsage["init"].(float64))
			e.heapMemoryUsage.WithLabelValues("max").Set(heapMemoryUsage["max"].(float64))
			e.heapMemoryUsage.WithLabelValues("used").Set(heapMemoryUsage["used"].(float64))
		}

		/*
		   {
		       "name": "Hadoop:service=NameNode,name=RpcActivityForPort8020",
		       "modelerType": "RpcActivityForPort8020",
		       "tag.port": "8020",
		       "tag.Context": "rpc",
		       "tag.NumOpenConnectionsPerUser": "{\"hive\":11,\"manas\":3,\"ossuser\":197,\"spark\":2,\"ambari-qa\":4,\"kafka\":1,\"hdfs\":53,\"yarn\":51,\"hbase\":50,\"mapred\":1}",
		       "tag.Hostname": "osb002.example.com",
		       "ReceivedBytes": 1505609759776,
		       "SentBytes": 4366768779986,
		       "RpcQueueTimeNumOps": 6291228413,
		       "RpcQueueTimeAvgTime": 0.02962496060510558,
		       "RpcProcessingTimeNumOps": 6291228413,
		       "RpcProcessingTimeAvgTime": 0.12858493539237315,
		       "RpcAuthenticationFailures": 638766,
		       "RpcAuthenticationSuccesses": 49398112,
		       "RpcAuthorizationFailures": 0,
		       "RpcAuthorizationSuccesses": 49397832,
		       "RpcClientBackoff": 0,
		       "RpcSlowCalls": 0,
		       "NumOpenConnections": 373,
		       "CallQueueLength": 0
		   },
		*/
		if strings.HasPrefix(nameDataMap["modelerType"].(string), "RpcActivityForPort") {

			port := nameDataMap["tag.port"].(string)

			e.RpcReceivedBytes.WithLabelValues(port).Set(nameDataMap["ReceivedBytes"].(float64))
			e.RpcSentBytes.WithLabelValues(port).Set(nameDataMap["SentBytes"].(float64))
			e.RpcQueueTimeNumOps.WithLabelValues(port, "QueueTime").Set(nameDataMap["RpcQueueTimeNumOps"].(float64))
			e.RpcAvgTime.WithLabelValues(port, "RpcQueueTime").Set(nameDataMap["RpcQueueTimeAvgTime"].(float64))
			e.RpcAvgTime.WithLabelValues(port, "RpcProcessingTime").Set(nameDataMap["RpcProcessingTimeAvgTime"].(float64))
			e.RpcNumOpenConnections.WithLabelValues(port).Set(nameDataMap["NumOpenConnections"].(float64))
			e.RpcCallQueueLength.WithLabelValues(port).Set(nameDataMap["CallQueueLength"].(float64))
		}
	}

	e.MissingBlocks.Collect(ch)
	e.UnderReplicatedBlocks.Collect(ch)
	e.Capacity.Collect(ch)
	e.BlocksTotal.Collect(ch)
	e.FilesTotal.Collect(ch)
	e.CorruptBlocks.Collect(ch)
	e.ExcessBlocks.Collect(ch)
	e.StaleDataNodes.Collect(ch)
	e.GcCount.Collect(ch)
	e.GcTime.Collect(ch)
	e.heapMemoryUsage.Collect(ch)
	e.lastHATransitionTime.Collect(ch)
	e.HAState.Collect(ch)
	e.RpcReceivedBytes.Collect(ch)
	e.RpcSentBytes.Collect(ch)
	e.RpcQueueTimeNumOps.Collect(ch)
	e.RpcAvgTime.Collect(ch)
	e.RpcNumOpenConnections.Collect(ch)
	e.RpcCallQueueLength.Collect(ch)
}

func main() {

	flag.Parse()

	exporter := NewExporter(*namenodeJmxUrl, *keytabPath, *principal)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
        <head><title>NameNode Exporter</title></head>
        <body>
        <h1>NameNode Exporter</h1>
        <p><a href="` + *metricsPath + `">Metrics</a></p>
        </body>
        </html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

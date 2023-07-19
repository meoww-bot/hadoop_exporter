# Hadoop Exporter for Prometheus
Exports hadoop metrics via HTTP for Prometheus consumption.


## Build

How to build
```
go mod tidy
make build
```

or build individual exporter
```
make build-namenode
make build-resourcemanager
make build-journalnode 
make build-datanode
```

## Help

Help on flags of namenode_exporter:
```
-krb5.keytab.path string
    	Kerberos keytab file path
-krb5.principal string
    	Principal (admin@EXAMPLE.COM)
-namenode.jmx.url string
    Hadoop JMX URL. (default "http://nn01.example.com:50070/jmx")
-web.listen-address string
    Address on which to expose metrics and web interface. (default ":9070")
-web.telemetry-path string
    Path under which to expose metrics. (default "/metrics")
```

Help on flags of datanode_exporter:
```
-datanode.jmx.url string
    Hadoop JMX URL. (default "http://localhost:50075/jmx")
-web.listen-address string
    Address on which to expose metrics and web interface. (default ":9070")
-web.telemetry-path string
    Path under which to expose metrics. (default "/metrics")
```

Help on flags of resourcemanager_exporter:
```
-resourcemanager.url string
    Hadoop ResourceManager URL. (default "http://localhost:8088")
-web.listen-address string
    Address on which to expose metrics and web interface. (default ":9088")
-web.telemetry-path string
    Path under which to expose metrics. (default "/metrics")
```

Help on flags of journalnode_exporter:
```
-journalnodeJmxUrl.url string
    Hadoop ResourceManager URL. (default "http://localhost:8088")
-web.listen-address string
    Address on which to expose metrics and web interface. (default ":9088")
-web.telemetry-path string
    Path under which to expose metrics. (default "/metrics")
```

## Metrics Map

### NameNode

#### Hadoop:service=NameNode,name=FSNamesystem

|Jmx Metric|Prometheus Metric|Description|Chinese Description|
|-|-|-|-|
|MissingBlocks|hdfs_namenode_fsname_system_missing_blocks|Current number of missing blocks|
|UnderReplicatedBlocks|hdfs_namenode_fsname_system_under_replicated_blocks|Current number of blocks under replicated 
|CapacityTotal|hdfs_namenode_fsname_system_capacity_bytes{mode="Total"}|Current raw capacity of DataNodes in bytes
|CapacityUsed|hdfs_namenode_fsname_system_capacity_bytes{mode="Used"}|Current used capacity across all DataNodes in bytes
|CapacityRemaining|hdfs_namenode_fsname_system_capacity_bytes{mode="Remaining"}|Current remaining capacity in bytes
|CapacityUsedNonDFS|hdfs_namenode_fsname_system_capacity_bytes{mode="UsedNonDFS"}|Current space used by DataNodes for non DFS purposes in bytes
|BlocksTotal|hdfs_namenode_fsname_system_blocks_total|Current number of allocated blocks in the system
|FilesTotal|hdfs_namenode_fsname_system_files_total|Current number of files and directories
|CorruptBlocks|hdfs_namenode_fsname_system_corrupt_blocks|Current number of blocks with corrupt replicas
|ExcessBlocks|hdfs_namenode_fsname_system_excess_blocks|Current number of excess blocks
|StaleDataNodes|hdfs_namenode_fsname_system_stale_datanodes|Current number of DataNodes marked stale due to delayed heartbeat
|tag.HAState|hdfs_namenode_fsname_system_hastate|(HA-only) Current state of the NameNode: initializing or active or standby or stopping state |


#### Hadoop:service=NameNode,name=JvmMetrics

|Jmx Metric|Prometheus Metric|Description|Chinese Description|
|-|-|-|-|
|GcCountParNew|hdfs_namenode_jvm_gc_count{type="ParNew"}|ParNew GC count
|GcCountConcurrentMarkSweep|hdfs_namenode_jvm_gc_count{type="ConcurrentMarkSweep"}|ConcurrentMarkSweep GC count
|GcTimeMillisParNew|hdfs_namenode_jvm_gc_time_milliseconds{type="ParNew"}|ParNew GC time in milliseconds
|GcTimeMillisConcurrentMarkSweep|hdfs_namenode_jvm_gc_time_milliseconds{type="ConcurrentMarkSweep"}|ConcurrentMarkSweep GC time in milliseconds


#### java.lang:type=Memory

|Jmx Metric|Prometheus Metric|Description|Chinese Description|
|-|-|-|-|
|HeapMemoryUsage{committed}|hdfs_namenode_mem_heap_memory_usage_bytes{mode="committed"}|
|HeapMemoryUsage{init}|hdfs_namenode_mem_heap_memory_usage_bytes{mode="init"}|
|HeapMemoryUsage{max}|hdfs_namenode_mem_heap_memory_usage_bytes{mode="max"}|
|HeapMemoryUsage{used}|hdfs_namenode_mem_heap_memory_usage_bytes{mode="used"}|

#### Hadoop:service=NameNode,name=NameNodeStatus

|Jmx Metric|Prometheus Metric|Description|Chinese Description|
|-|-|-|-|
|LastHATransitionTime|hdfs_namenode_namenode_status_last_ha_transition_time|


####  Hadoop:service=NameNode,name=RpcActivityForPort8020/8060

|Jmx Metric|Prometheus Metric|Description|Chinese Description|
|-|-|-|-|
|ReceivedBytes|hdfs_namenode_rpc_received_bytes|Total number of received bytes
|SentBytes|hdfs_namenode_rpc_sent_bytes|Total number of sent bytes
|RpcQueueTimeNumOps|hdfs_namenode_rpc_call_count{method="QueueTime"}|Total number of RPC calls 
|RpcQueueTimeAvgTime|hdfs_namenode_rpc_avg_time_milliseconds{method="RpcQueueTime"}|Average queue time in milliseconds 
|RpcProcessingTimeAvgTime|hdfs_namenode_rpc_avg_time_milliseconds{method="RpcProcessingTime"}|Average Processing time in milliseconds
|NumOpenConnections|hdfs_namenode_rpc_open_connections_count|Current number of open connections
|CallQueueLength|hdfs_namenode_rpc_call_queue_length|Current length of the call queue




# Requirements
golang 1.20

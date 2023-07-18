# Hadoop Exporter for Prometheus
Exports hadoop metrics via HTTP for Prometheus consumption.

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

Help on flags of namenode_exporter:
```
-krb5.keytabpath string
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

# Requirements
golang 1.20

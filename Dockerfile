FROM golang:1.8.3
MAINTAINER Nicholas Laferriere ops@tamr.com

RUN mkdir -p /go/src/github.com/Datatamer/hdfs_exporter

ADD . /go/src/github.com/Datatamer/hdfs_exporter

## Build Code
RUN cd /go/bin && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/namenode/namenode_exporter.go && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/resourcemanager/resourcemanager_exporter.go && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/journalnode/journalnode_exporter.go && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/datanode/datanode_exporter.go

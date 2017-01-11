FROM golang:1.6.2
MAINTAINER Nicholas Laferriere dev@tamr.com

RUN mkdir -p /go/src/github.com/Datatamer/hdfs_exporter

ADD . /go/src/github.com/Datatamer/hdfs_exporter

## Build Code
RUN cd /go/bin && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/namenode/namenode_exporter.go && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/resourcemanager/resourcemanager_exporter.go && \
    go build /go/src/github.com/Datatamer/hdfs_exporter/journalnode/journalnode_exporter.go

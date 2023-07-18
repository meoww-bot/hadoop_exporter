SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '-CHANGES' || true)
VERSION := $(shell cat ./VERSION)
DOCKER_REPO := quay.io/tamr/hdfs_exporter
DOCKER_TAG := $(shell cat ./VERSION)


.PHONY: \
	clean \
	test \
	vet \
	lint \
	build-namenode \
	build-resourcemanager \
	build-journalnode \
	build

all: fmt vet build

test:
	go test -v ./

vet:
	go vet -v ./

lint:
	golint ./

style:
	gofmt -d ./

build-namenode:
	go fmt ./namenode
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/namenode_exporter ./namenode/namenode_exporter.go

build-resourcemanager:
	go fmt ./resourcemanager
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/resourcemanager_exporter ./resourcemanager/resourcemanager_exporter.go

build-journalnode:
	go fmt ./journalnode
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/journalnode_exporter ./journalnode/journalnode_exporter.go

build-datanode:
	go fmt ./datanode
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/datanode_exporter ./datanode/datanode_exporter.go

build: build-namenode build-resourcemanager build-journalnode build-datanode

docker-build:
	echo "docker build tag: $(DOCKER_REPO):$(DOCKER_TAG)$(CHANGES)"
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG)$(CHANGES) .

SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '-CHANGES' || true)
VERSION := $(shell cat ./VERSION)
DOCKER_REPO := quay.io/tamr/hdfs_exporter
DOCKER_TAG := $(shell cat ./VERSION)


.PHONY: \
	clean \
	clean-vendor \
	deps \
	test \
	vet \
	lint \
	build-namenode \
	build-resourcemanager \
	build-journalnode \
	build

all: fmt vet build

clean-vendor:
	rm -rf ./vedor/

clean: clean-vendor

deps:
	glide install

test: deps
	go test -v ./

vet:
	go vet -v ./

lint:
	golint ./

style:
	gofmt -d ./

build-namenode: deps
	go fmt ./namenode
	go build -o bin/namenode_exporter ./namenode/namenode_exporter.go

build-resourcemanager: deps
	go fmt ./resourcemanager
	go build -o bin/resourcemanager_exporter ./resourcemanager/resourcemanager_exporter.go

build-journalnode: deps
	go fmt ./journalnode
	go build -o bin/journalnode_exporter ./journalnode/journalnode_exporter.go

build: build-namenode build-resourcemanager build-journalnode

docker-build:
	echo "docker build tag: $(DOCKER_REPO):$(DOCKER_TAG)$(CHANGES)"
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG)$(CHANGES) .

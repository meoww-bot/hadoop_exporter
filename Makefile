SHELL := /bin/bash

REV := $(shell git rev-parse HEAD)
CHANGES := $(shell test -n "$$(git status --porcelain)" && echo '-CHANGES' || true)
VERSION := $(shell )
DOCKER_REPO := quay.io/tamr/prom
DOCKER_TAG := $(shell cat ./version)


.PHONY: \
	clean \
	clean-vendor \
	deps \
	test \
	vet \
	lint \
	fmt \
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

fmt:
	go fmt ./

build: fmt deps
	go build namenode/namenode_exporter.go
	go build resourcemanager/resourcemanager_exporter.go

docker-build:
	docker build -t $(DOCKER_REPO):$(DOCKER_TAG)$(CHANGES) .

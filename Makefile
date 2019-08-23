.PHONY: *

.DEFAULT_GOAL := build

DIST=dist

dependencies:
	# https://github.com/golang/go/issues/26610
	# This downloads dependencies listed in the go.mod file and not the source
	# code. Not too useful while building locally, because go build will also
	# fetch dependencies, but useful in a Dockerfile to cache dependencies in a
	# separate layer
	go mod download

build: build-linux-amd64 build-darwin-amd64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/s3s2-linux-amd64 -v .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/s3s2-darwin-amd64 -v .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/s3s2-windows-amd64.exe -v .

publish-nexus: build
	# Nothing yet

clean:
	rm $(DIST)/*

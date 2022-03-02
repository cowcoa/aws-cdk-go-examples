.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go env -w CGO_ENABLED="0"
	go vet ./...
	shadow ./...
.PHONY:vet

build: vet
	go build
.PHONY:build

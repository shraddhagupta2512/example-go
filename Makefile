.PHONY: build run run_mqtt test

build:
	CGO_ENABLED=1 go build -o cmd/example-go ./cmd

run:
	cd cmd && ./example-go -cfg=./res/config.json


run_mqtt:
	cd cmd && ./example-go -cfg=./res/config.json

test:
	go test ./... -coverprofile=coverage.out ./...
	go vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

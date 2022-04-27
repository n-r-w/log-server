.PHONY: build test run runbuild proto rebuild tidy

build:
	go build -v -o . ./cmd/logserver

rebuild:
	go build -a -v -o . ./cmd/logserver

test:
	go test -v -race -timeout 30s ./...

run:
	go run ./cmd/logserver

runbuild:
	./bin/logserver

tidy:
	go mod tidy

proto:
	protoc --proto_path=./api/proto --go_out=./api/schema ./api/proto/log.proto

.DEFAULT_GOAL := run
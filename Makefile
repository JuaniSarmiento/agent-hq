.PHONY: build build-mcp install run clean test

build:
	go build -o bin/agenthq ./cmd/agenthq

build-mcp:
	go build -o bin/agenthq-mcp ./cmd/agenthq-mcp

install: build build-mcp
	cp bin/agenthq ~/.local/bin/
	cp bin/agenthq-mcp ~/.local/bin/

run:
	go run ./cmd/agenthq

clean:
	rm -rf bin/

test:
	go test ./...

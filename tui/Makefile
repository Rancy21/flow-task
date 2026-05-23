.DEFAULT_GOAL := build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build ./...

.PHONY: clean
clean:
	go clean ./...

.PHONY: run
run:
	go run ./cmd/

.PHONY: binary
binary:
	go build -ldflags="-s -w" -o /home/larryck/.local/bin/flowtask ./cmd/main.go

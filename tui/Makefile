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

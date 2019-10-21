.PHONY: all build test test-race test-cover lint docker

all: build
	. secrets/.env-test && ./bin/gateway

build:
	mkdir -p bin
	go build -o bin/gateway \
        -ldflags "-X main.version=$${VERSION:-$$(git describe --tags --always --dirty)}" \
        ./cmd/gateway/main.go

test:
	. secrets/.env-test && go test -v ./...

test-race:
	. secrets/.env-test && go test -v -race -cover ./...

test-cover:
	rm -f all.coverage.out
	. secrets/.env-test  && go test -race -coverprofile=all.coverage.out -coverpkg=./... $$(go list ./...|grep -v cmd)

test-docker:
	./test.sh

lint:
	docker pull golangci/golangci-lint:latest
	docker run -v`pwd`:/workspace -w /workspace \
        golangci/golangci-lint:latest golangci-lint run ./...

.PHONY: all build test test-race test-cover lint

all: build
	. secrets/.env-test && ./bin/driver-location

build:
	mkdir -p bin
	go build -o bin/driver-location \
        -ldflags "-X main.version=$${VERSION:-$$(git describe --tags --always --dirty)}" \
        ./cmd/driver-location/main.go

test:
	. secrets/.env-test && go test ./...

test-race:
	. secrets/.env-test && go test -race -cover ./...

test-cover:
	rm -f all.coverage.out
	. secrets/.env-test  && go test -race -coverprofile=all.coverage.out -coverpkg=./... $$(go list ./...|grep -v cmd)

test-docker:
	. secrets/.env-test && ../test.sh

lint:
	docker pull golangci/golangci-lint:latest
	docker run -v`pwd`:/workspace -w /workspace \
        golangci/golangci-lint:latest golangci-lint run ./...

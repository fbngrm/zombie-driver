.PHONY: all build test test-race test-cover lint

all: build

build:
	mkdir -p bin
	go build -o bin/zombie-driver \
        -ldflags "-X main.version=$${VERSION:-$$(git describe --tags --always --dirty)}" \
        ./cmd/zombie-driver/main.go

test:
	../test.sh -timeout=1m

test-race:
	../test.sh -race -timeout=2m

test-cover:
	rm -f all.coverage.out
	../test.sh -race -timeout=2m \
		-coverprofile=all.coverage.out \
		-coverpkg=./... $$(go list ./...|grep -v cmd)

lint:
	docker pull golangci/golangci-lint:latest
	docker run -v`pwd`:/workspace -w /workspace \
        golangci/golangci-lint:latest golangci-lint run ./...

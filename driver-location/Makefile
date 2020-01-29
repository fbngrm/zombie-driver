.PHONY: all build test test-race test-cover

all: build

build:
	mkdir -p bin
	go build -o bin/driver-location \
        -ldflags "-X main.version=$${VERSION:-$$(git describe --tags --always --dirty)}" \
        ./cmd/driver-location/main.go

test:
	../test.sh -timeout=1m

test-race:
	../test.sh -race -timeout=2m

test-cover:
	rm -f all.coverage.out
	../test.sh -race -timeout=2m \
		-coverprofile=all.coverage.out \
		-coverpkg=./... $$(go list ./...|grep -v cmd)

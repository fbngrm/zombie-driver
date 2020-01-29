.PHONY: all build up test test-race test-cover lint

export HOST_NAME=void

all: build

build:
	make -C driver-location/ all
	make -C gateway/ all
	make -C zombie-driver/ all

up:
	docker-compose up

test:
	./test.sh -timeout=1m

test-race:
	./test.sh -race -timeout=2m

test-cover:
	rm -f all.coverage.out
	./test.sh -race -timeout=2m \
		-coverprofile=all.coverage.out \
		-coverpkg=./... $$(go list ./...|grep -v cmd)

lint:
	docker pull golangci/golangci-lint:latest
	docker run -v`pwd`:/workspace -w /workspace \
        golangci/golangci-lint:latest golangci-lint run ./...

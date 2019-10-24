.PHONY: all build driver-location zombie-driver gateway docker test
export HOST_NAME=void

all:

build:
	make -C ./driver-location build
	make -C ./gateway build
	make -C ./zombie-driver build

docker:
	docker-compose up

test:
	./test.sh

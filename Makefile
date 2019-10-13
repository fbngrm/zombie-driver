.PHONY: all test docker driver-location zombie-driver gateway

docker:
	export HOST_NAME=void
	docker-compose up

all:
	make -j5 docker driver-location gateway zombie-driver

driver-location:
	make -C ./driver-location all

gateway:
	make -C ./gateway all

zombie-driver:
	make -C ./zombie-driver all

test:
	make -C ./driver-location test
	make -C ./gateway test
	make -C ./zombie-driver test

SHELL := bash

.PHONY : build-static

IMAGE_TAG = coriolis-ovm-exporter-builder

build-static:
	docker build --tag $(IMAGE_TAG) .
	docker run --rm -v $(PWD):/build/coriolis-ovm-exporter $(IMAGE_TAG) /build-static.sh

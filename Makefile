IMAGE_REPO ?= scheduler-extender-demo
IMAGE_TAG ?= v0.0.1

.PHONY: all
all: build build-images

.PHONY: build
build: build-scheduler

.PHONY: build-scheduler
build-scheduler:
	 go build -o bin/kube-scheduler cmd/scheduler-extender/main.go

.PHONY: build-images
build-images:
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .
VERSION := 0.1.0
DOCKER_USER := ae6rt

# No changes required below this line ------------
NAME := bctool
ARCH := amd64
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildInfo=Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)"'
SOURCEC:=$(shell go list ./... | grep -v vendor/)

IMAGE := $(DOCKER_USER)/decap-build-base

all: binaries

lint: 
	@go vet 
	@go fmt
	@errcheck

test:  lint
	@go test ${SOURCES}

binaries: test
	GOOS=darwin GOARCH=$(ARCH) go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH)
	GOOS=linux GOARCH=$(ARCH) go build -ldflags $(LD_FLAGS) -o $(NAME)-linux-$(ARCH)

image: binaries
	docker build -t $(IMAGE):latest .  
	docker tag $(IMAGE):latest ${IMAGE}:${VERSION}

push: image
	docker push $(IMAGE):latest
	docker push $(IMAGE):${VERSION}

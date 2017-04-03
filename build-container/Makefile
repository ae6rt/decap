NAME := bctool
ARCH := amd64
VERSION := 0.2
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildInfo=Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)"'
SOURCEC:=$(shell go list ./... | grep -v vendor/)

# Change this to your DockerHub user when developing decap
DOCKER_USER := ae6rt

# No changes required below this line ------------

IMAGE := $(DOCKER_USER)/decap-build-base:latest

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

container: binaries
	docker build -t $(IMAGE) .  

push: container
	docker push $(IMAGE)

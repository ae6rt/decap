NAME := bctool
ARCH := amd64
VERSION := 0.1
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildInfo=Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)"'

# Change this to your DockerHub user when developing decap
DOCKER_USER := ae6rt

# No changes required below this line ------------

IMAGE := $(DOCKER_USER)/decap-build-base:latest

all: binaries

test: 
	godep go test -v ./...

binaries: test
	GOOS=darwin GOARCH=$(ARCH) godep go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH)
	GOOS=linux GOARCH=$(ARCH) godep go build -ldflags $(LD_FLAGS) -o $(NAME)-linux-$(ARCH)

container:
	docker --tlsverify=false build -t $(IMAGE) .  

push: container
	docker --tlsverify=false push $(IMAGE)

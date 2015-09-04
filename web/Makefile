NAME := decap
ARCH := amd64
VERSION := 0.1
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildInfo=Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)"'

all: clean binaries 

test:
	godep go test -v

binaries: test 
	CGO_ENABLED=0 godep go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH) 
	GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=0 godep go build -ldflags $(LD_FLAGS) -o $(NAME)-linux-$(ARCH) 

container: binaries
	docker --tlsverify=false build -t ae6rt/decap:latest .

push: container
	docker --tlsverify=false push  ae6rt/decap:latest 

clean: 
	godep go clean

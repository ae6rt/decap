# Change this to your DockerHub user when developing decap
DOCKER_USER := ae6rt

# No changes required below this line ------------

NAME := decap
ARCH := amd64
VERSION := 0.1
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildVersion=$(VERSION)" -X "main.buildCommit=$(COMMIT_ID)" -X "main.buildDate=$(DATE)" -X "main.buildGoSDK=$(SDK_INFO)"'
IMAGE=$(DOCKER_USER)/decap:latest

all: clean binaries 

test:
	godep go test -v

osx: test
	CGO_ENABLED=0 godep go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH) 

binaries: test 
	CGO_ENABLED=0 godep go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH) 
	GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=0 godep go build -ldflags $(LD_FLAGS) -o $(NAME)-linux-$(ARCH) 

container: binaries
	docker --tlsverify=false build -t $(IMAGE) .

push: container
	docker --tlsverify=false push  $(IMAGE)

clean: 
	godep go clean

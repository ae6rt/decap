# Change this to your DockerHub user when developing decap
DOCKER_USER := ae6rt

# No changes required below this line ------------

NAME := decap
ARCH := amd64
VERSION := 0.2
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo $$COMMIT_ID)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildVersion=$(VERSION)" -X "main.buildCommit=$(COMMIT_ID)" -X "main.buildDate=$(DATE)" -X "main.buildGoSDK=$(SDK_INFO)"'
IMAGE=$(DOCKER_USER)/decap:latest
SOURCES:=$(shell go list ./... | grep -v vendor/)

all: clean binaries 

lint:
	@echo Linting...
	@go fmt ${SOURCES}
	@go vet ${SOURCES}
	@type errcheck >/dev/null 2>&1 || go get github.com/kisielk/errcheck
	@errcheck ${SOURCES}

test: lint
	@echo Running tests...
	@go test ${SOURCES}

coverage:
	@mkdir -p reports/
	gocov test ${SOURCES} | gocov-html  > reports/coverage.html

binary-linux: 
	@mkdir -p build/dist/linux
	@echo Build binary
	GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags $(LD_FLAGS) -o build/dist/linux/$(NAME)

app: 
	@mkdir -p build/dist/linux
	@echo Build binary
	cd cmd && GOOS=linux GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags $(LD_FLAGS) -o app && cd ..

binary-darwin: 
	@mkdir -p build/dist/darwin
	GOOS=darwin CGO_ENABLED=0 go build -ldflags $(LD_FLAGS) -o $(NAME)-darwin-$(ARCH) 

binaries: test binary-linux binary-darwin

image: binary-linux
	docker build -t $(IMAGE) .

push: image
	docker push  $(IMAGE)

clean: 
	@go clean

.PHONY: all lint test osx binaries container push clean app

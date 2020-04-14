GOCMD ?= go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
BINARY_NAME = livego
BINARY_UNIX = $(BINARY_NAME)_unix

DOCKER_ACC ?= gwuhaolin
DOCKER_REPO ?= livego

TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

default: all

all: test build dockerize
build:
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags="-X main.VERSION=$(TAG)"

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run: build
	./$(BINARY_NAME)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

dockerize:
	docker build -t $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG) .
	docker push $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG)

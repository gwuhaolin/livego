GO_BIN ?= go
GO_BUILD = $(GO_BIN) build
GO_CLEAN = $(GO_BIN) clean
GO_TEST = $(GO_BIN) test
GO_GET = $(GO_BIN) get
BINARY_NAME = livego
BINARY_UNIX = $(BINARY_NAME)_unix

DOCKER_ACC ?= gwuhaolin
DOCKER_REPO ?= livego

TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

tidy:
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
else
	echo skipping go mod tidy
endif

run: binary
	./$(BINARY_NAME)

build:
	pkger
	$(GO_BUILD) -o $(BINARY_NAME) -v -ldflags="-X main.VERSION=$(TAG)"
	make tidy

test:
	pkger
	$(GO_TEST) -tags ${TAGS} -cover ./...
	pkger
	make tidy

lint:
	golangci-lint --vendor ./... --deadline=1m --skip=internal

## Build WebUI Docker image
build-webui-image:
	docker build -t livego-webui -f webui/Dockerfile webui

generate-webui: build-webui-image
	if [ ! -d "static" ]; then \
		mkdir -p static; \
		docker run --rm -v "$$PWD/static":'/src/webui/build' livego-webui npm run build; \
		docker run --rm -v "$$PWD/static":'/src/webui/build' livego-webui chown -R $(shell id -u):$(shell id -g) ./build; \
		echo 'For more informations show `webui/readme.md`' > $$PWD/static/DONT-EDIT-FILES-IN-THIS-DIRECTORY.md; \
	fi

dockerize:
	docker build -t $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG) .
	docker push $(DOCKER_ACC)/$(DOCKER_REPO):$(TAG)

binary: generate-webui
	make build

default: binary

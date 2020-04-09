GO_BIN ?= go

tidy:
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
else
	echo skipping go mod tidy
endif

build:
	pkger
	$(GO_BIN) build -v .
	make tidy

test:
	pkger
	$(GO_BIN) test -tags ${TAGS} -cover ./...
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
	docker build -t gwuhaolin:livego .

binary: generate-webui
	make build

default: binary

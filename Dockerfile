# WEBUI
FROM node:13.12 as webui

ENV WEBUI_DIR /src/webui
RUN mkdir -p $WEBUI_DIR

COPY ./webui/ $WEBUI_DIR/

WORKDIR $WEBUI_DIR

RUN npm install
RUN npm run build

# BUILD
FROM golang:latest as gobuild

RUN go get github.com/markbates/pkger/cmd/pkger

WORKDIR /go/src/github.com/gwuhaolin/livego

# Download go modules
COPY go.mod .
COPY go.sum .
RUN GO111MODULE=on GOPROXY=https://proxy.golang.org go mod download

ENV REPO github.com/gwuhaolin/livego

COPY . /go/src/github.com/gwuhaolin/livego

RUN rm -rf /go/src/github.com/gwuhaolin/livego/static/
COPY --from=webui /src/webui/static/ /go/src/github.com/gwuhaolin/livego/static/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o livego .

## IMAGE
FROM alpine:3.10

COPY --from=gobuild /go/src/github.com/gwuhaolin/livego/livego /

VOLUME ["/tmp"]

ENV RTMP_PORT 1935
ENV HTTP_FLV_PORT 7001
ENV HLS_PORT 7002
ENV HTTP_OPERATION_PORT 8090

EXPOSE ${RTMP_PORT}
EXPOSE ${HTTP_FLV_PORT}
EXPOSE ${HLS_PORT}
EXPOSE ${HTTP_OPERATION_PORT}

ENTRYPOINT ["/livego"]

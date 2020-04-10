![Test](https://github.com/gwuhaolin/livego/workflows/Test/badge.svg)

<img src='./logo.png' width='130px' height='50px'/>

# livego
Simple and efficient live broadcast server:
- Very simple to install and use;
- Pure Golang, high performance, cross-platform;
- Support commonly used transmission protocols, file formats, encoding formats;

#### Supported transport protocols
- RTMP
- AMF
- HLS
- HTTP-FLV

#### Supported container formats
- FLV
- TS

#### Supported encoding formats
- H264
- AAC
- sMP3

## Installation
After directly downloading the compiled [binary file](https://github.com/gwuhaolin/livego/releases), execute it on the command line.

#### Boot from Docker
Run `docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -d --name livego docker.pkg.github.com/gwuhaolin/livego:latest` to start

#### Compile from source
1. Download the source code `git clone https://github.com/gwuhaolin/livego.git`
2. Go to the livego directory and execute `go build`

## Use
2. Start the service: execute the livego binary file to start the livego service;
3. Upstream push: Push the video stream to `rtmp://localhost:1935/live/movie` through the` RTMP` protocol, for example, use `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/movie` push;
4. Downstream playback: The following three playback protocols are supported, and the playback address is as follows:
    -`RTMP`:`rtmp://localhost:1935/live/movie`
    -`FLV`:`http://127.0.0.1:7001/live/movie.flv`
    -`HLS`:`http://127.0.0.1:7002/live/movie.m3u8`

### [Use with flv.js](https://github.com/gwuhaolin/blog/issues/3)

Interested in Golang? Please see [Golang Chinese Learning Materials Summary](http://go.wuhaolin.cn/)

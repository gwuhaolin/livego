<p align='center'>
    <img src='./logo.png' width='200px' height='80px'/>
</p>

[![Test](https://github.com/gwuhaolin/livego/workflows/Test/badge.svg)](https://github.com/gwuhaolin/livego/actions?query=workflow%3ATest)

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
- MP3

## Installation
After directly downloading the compiled [binary file](https://github.com/gwuhaolin/livego/releases), execute it on the command line.

#### Boot from Docker
Run `docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -d --name livego gwuhaolin/livego` to start

#### Compile from source
1. Download the source code `git clone https://github.com/gwuhaolin/livego.git`
2. Go to the livego directory and execute `go build` or `make build`

## Use
```bash
./livego  -h
Usage of ./livego:
      --api_addr string       HTTP manage interface server listen address (default ":8090")
      --config_file string    configure filename (default "livego.yaml")
      --flv_dir string        output flv file at flvDir/APP/KEY_TIME.flv (default "tmp")
      --gop_num int           gop num (default 1)
      --hls_addr string       HLS server listen address (default ":7002")
      --httpflv_addr string   HTTP-FLV server listen address (default ":7001")
      --level string          Log level (default "info")
      --read_timeout int      read time out (default 10)
      --rtmp_addr string      RTMP server listen address (default ":1935")
      --write_timeout int     write time out (default 10)
```
2. Start the service: execute the livego binary file or `make run` to start the livego service;
3. Get a channelkey `curl http://localhost:8090/control/get?room=movie` and copy data like your channelkey.
4. Upstream push: Push the video stream to `rtmp://localhost:1935/live/movie`(`rtmp://localhost:1935/{appname}/{channelkey}`) through the` RTMP` protocol, for example, use `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/movie` push;
5. Downstream playback: The following three playback protocols are supported, and the playback address is as follows:
    -`RTMP`:`rtmp://localhost:1935/live/movie`
    -`FLV`:`http://127.0.0.1:7001/live/movie.flv`
    -`HLS`:`http://127.0.0.1:7002/live/movie.m3u8`

### [Use with flv.js](https://github.com/gwuhaolin/blog/issues/3)

Interested in Golang? Please see [Golang Chinese Learning Materials Summary](http://go.wuhaolin.cn/)

# livego
live streaming server write in pure go, simple efficient and can run in any platform.

## Support
#### Transport protocol
- [x] RTMP
- [x] AMF
- [x] HLS
- [x] HTTP-FLV
- [ ] WebSocket-FLV
#### File container
- [x] FLV
- [x] TS
#### AV coder
- [x] H264
- [x] AAC
- [x] MP3

## Install
### Download Bin
[releases](https://github.com/gwuhaolin/livego/releases)

### Docker
TODO

### Install System Service
- Mac

### Build From Source code
1. run `git clone https://github.com/gwuhaolin/livego.git`
2. cd to livego dir then run `go build`

## Use
2. run  `livego` to start livego server
3. push `RTMP` stream to `rtmp://localhost:1935/live/movie`, eg use `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/movie`
4. play stream use [VLC](http://www.videolan.org/vlc/index.html) or other players 
    - play `RTMP` from `rtmp://localhost:1935/live/movie`
    - play `FLV` from `http://127.0.0.1:8081/live/movie.flv`
    - play `HLS` from `http://127.0.0.1:8082/live/movie.m3u8`

# livego - a streaming server in pure go

## Feature
- write in pure golang, can run in any platform
- for live streaming
- support `RTMP` and `FLV` `HLS` over HTTP

## Use
1. run `git clone `
2. run  `go run main.go` to start livego server
3. push `RTMP` stream to `rtmp://localhost:1935/live/movie`, eg use `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/live/movie`
4. play stream use [VLC](http://www.videolan.org/vlc/index.html) or other players 
    - play `RTMP` from `rtmp://localhost:1935/live/movie`
    - play `FLV` from `http://127.0.0.1:8081/live/movie.flv`
    - play `HLS` from `http://127.0.0.1:8082/live/movie.m3u8`

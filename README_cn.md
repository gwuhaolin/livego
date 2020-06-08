<p align='center'>
    <img src='./logo.png' width='200px' height='80px'/>
</p>

[![Test](https://github.com/gwuhaolin/livego/workflows/Test/badge.svg)](https://github.com/gwuhaolin/livego/actions?query=workflow%3ATest)
[![Release](https://github.com/gwuhaolin/livego/workflows/Release/badge.svg)](https://github.com/gwuhaolin/livego/actions?query=workflow%3ARelease)

简单高效的直播服务器：
- 安装和使用非常简单；
- 纯 Golang 编写，性能高，跨平台；
- 支持常用的传输协议、文件格式、编码格式；

#### 支持的传输协议
- RTMP
- AMF
- HLS
- HTTP-FLV

#### 支持的容器格式
- FLV
- TS

#### 支持的编码格式
- H264
- AAC
- MP3

## 安装
直接下载编译好的[二进制文件](https://github.com/gwuhaolin/livego/releases)后，在命令行中执行。

#### 从 Docker 启动
执行`docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -p 8090:8090 -d gwuhaolin/livego`启动

#### 从源码编译
1. 下载源码 `git clone https://github.com/gwuhaolin/livego.git`
2. 去 livego 目录中 执行 `go build`

## 使用
1. 启动服务：执行 `livego` 二进制文件启动 livego 服务；
2. 访问 `http://localhost:8090/control/get?room=movie` 获取一个房间的 channelkey(channelkey用于推流，movie用于播放).
3. 推流: 通过`RTMP`协议推送视频流到地址 `rtmp://localhost:1935/{appname}/{channelkey}` (appname默认是`live`), 例如： 使用 `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/{appname}/{channelkey}` 推流([下载demo flv](https://s3plus.meituan.net/v1/mss_7e425c4d9dcb4bb4918bbfa2779e6de1/mpack/default/demo.flv));
4. 播放: 支持多种播放协议，播放地址如下:
    - `RTMP`:`rtmp://localhost:1935/{appname}/movie`
    - `FLV`:`http://127.0.0.1:7001/{appname}/movie.flv`
    - `HLS`:`http://127.0.0.1:7002/{appname}/movie.m3u8`

所有配置项: 
```bash
./livego  -h
Usage of ./livego:
      --api_addr string       HTTP管理访问监听地址 (default ":8090")
      --config_file string    配置文件路径 (默认 "livego.yaml")
      --flv_dir string        输出的 flv 文件路径 flvDir/APP/KEY_TIME.flv (默认 "tmp")
      --gop_num int           gop 数量 (default 1)
      --hls_addr string       HLS 服务监听地址 (默认 ":7002")
      --hls_keep_after_end    Maintains the HLS after the stream ends
      --httpflv_addr string   HTTP-FLV server listen address (默认 ":7001")
      --level string          日志等级 (默认 "info")
      --read_timeout int      读超时时间 (默认 10)
      --rtmp_addr string      RTMP 服务监听地址 (默认 ":1935")
      --write_timeout int     写超时时间 (默认 10)
```

### [和 flv.js 搭配使用](https://github.com/gwuhaolin/blog/issues/3)

对Golang感兴趣？请看[Golang 中文学习资料汇总](http://go.wuhaolin.cn/)


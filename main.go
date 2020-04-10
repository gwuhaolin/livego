package main

import (
	"flag"
	"fmt"
	"livego/configure"
	"livego/protocol/api"
	"livego/protocol/hls"
	"livego/protocol/httpflv"
	"livego/protocol/rtmp"
	"net"
	"path"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

var VERSION = "master"

var (
	rtmpAddr       = flag.String("rtmp-addr", ":1935", "RTMP server listen address")
	httpFlvAddr    = flag.String("httpflv-addr", ":7001", "HTTP-FLV server listen address")
	hlsAddr        = flag.String("hls-addr", ":7002", "HLS server listen address")
	operaAddr      = flag.String("manage-addr", ":8090", "HTTP manage interface server listen address")
	configfilename = flag.String("config-file", "livego.json", "configure filename")
	levelLog       = flag.String("level", "info", "Log level")
)

func startHls() *hls.Server {
	hlsListen, err := net.Listen("tcp", *hlsAddr)
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("HLS server panic: ", r)
			}
		}()
		log.Info("HLS listen On ", *hlsAddr)
		hlsServer.Serve(hlsListen)
	}()
	return hlsServer
}

func startRtmp(stream *rtmp.RtmpStream, hlsServer *hls.Server) {
	rtmpListen, err := net.Listen("tcp", *rtmpAddr)
	if err != nil {
		log.Fatal(err)
	}

	var rtmpServer *rtmp.Server

	if hlsServer == nil {
		rtmpServer = rtmp.NewRtmpServer(stream, nil)
		log.Info("hls server disable....")
	} else {
		rtmpServer = rtmp.NewRtmpServer(stream, hlsServer)
		log.Info("hls server enable....")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("RTMP server panic: ", r)
		}
	}()
	log.Info("RTMP Listen On ", *rtmpAddr)
	rtmpServer.Serve(rtmpListen)
}

func startHTTPFlv(stream *rtmp.RtmpStream) {
	flvListen, err := net.Listen("tcp", *httpFlvAddr)
	if err != nil {
		log.Fatal(err)
	}

	hdlServer := httpflv.NewServer(stream)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("HTTP-FLV server panic: ", r)
			}
		}()
		log.Info("HTTP-FLV listen On ", *httpFlvAddr)
		hdlServer.Serve(flvListen)
	}()
}

func startAPI(stream *rtmp.RtmpStream) {
	if *operaAddr != "" {
		opListen, err := net.Listen("tcp", *operaAddr)
		if err != nil {
			log.Fatal(err)
		}
		opServer := api.NewServer(stream, *rtmpAddr)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("HTTP-API server panic: ", r)
				}
			}()
			log.Info("HTTP-API listen On ", *operaAddr)
			opServer.Serve(opListen)
		}()
	}
}

func InitLog() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})

	if l, err := log.ParseLevel(*levelLog); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("livego panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	// Log options
	InitLog()

	log.Infof(`
     _     _            ____       
    | |   (_)_   _____ / ___| ___  
    | |   | \ \ / / _ \ |  _ / _ \ 
    | |___| |\ V /  __/ |_| | (_) |
    |_____|_| \_/ \___|\____|\___/ 
        version: %s
	`, VERSION)
	configure.LoadConfig(*configfilename)

	stream := rtmp.NewRtmpStream()
	hlsServer := startHls()
	startHTTPFlv(stream)
	startAPI(stream)

	startRtmp(stream, hlsServer)
}
